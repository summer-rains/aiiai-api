package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"log"
	"one-api/common"
	"one-api/constant"
	"one-api/model"
	"one-api/service"
	"time"
)

func GetWxPayClient(ctx context.Context) *core.Client {
	if constant.WxPayMchId == "" || constant.WxPayApiV3Key == "" || constant.WxPaySerialNo == "" || constant.WxPayKeyPath == "" || constant.WxPayCertPath == "" {
		return nil
	}
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(constant.WxPayKeyPath)
	if err != nil {
		log.Fatal("load merchant private key error")
	}

	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(constant.WxPayMchId, constant.WxPaySerialNo, mchPrivateKey, constant.WxPayApiV3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		log.Fatalf("new wechat pay client err:%s", err)
		return nil
	}

	return client
}

func WxPayNative(c *gin.Context) {
	var req EpayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getMinTopup())})
		return
	}

	id := c.GetInt("id")
	group, err := model.CacheGetUserGroup(id)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getPayMoney(float64(req.Amount), group)
	if payMoney < 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	callBackAddress := service.GetCallbackAddress()
	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", id, tradeNo)

	amount := req.Amount
	if !common.DisplayInCurrencyEnabled {
		amount = amount / int(common.QuotaPerUnit)
	}
	topUp := &model.TopUp{
		UserId:     id,
		Amount:     amount,
		Money:      payMoney,
		TradeNo:    tradeNo,
		CreateTime: time.Now().Unix(),
		Status:     "pending",
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}
	//log.Printf("orderId: %d", topUp.Id)

	ctx := context.Background()
	client := GetWxPayClient(ctx)
	if client == nil {
		c.JSON(200, gin.H{"message": "error", "data": "当前管理员未配置支付信息"})
		return
	}

	svc := native.NativeApiService{Client: client}
	// 发送请求
	resp, _, err := svc.Prepay(ctx,
		native.PrepayRequest{
			Appid:       core.String(constant.WxPayAppId),
			Mchid:       core.String(constant.WxPayMchId),
			Description: core.String(fmt.Sprintf("%d的支付订单", id)),
			OutTradeNo:  core.String(tradeNo),
			Attach:      core.String(fmt.Sprintf("%d的支付订单", id)),
			NotifyUrl:   core.String(callBackAddress + "/api/user/wxpay/notify"),
			Amount: &native.Amount{
				Total: core.Int64(int64(payMoney * 100)),
			},
		},
	)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建微信支付订单支付失败！"})
		return
	}
	// 使用微信扫描 resp.code_url 对应的二维码，即可体验Native支付
	//log.Printf("status=%d resp=%s", result.Response.StatusCode, resp)

	// 放入redis
	common.RedisSet("topup_order::"+tradeNo, topUp.Status, time.Duration(6)*time.Minute)

	c.JSON(200, gin.H{"message": "success", "data": tradeNo, "url": resp.CodeUrl})
}

func WxPayNotify(c *gin.Context) {
	log.Printf("收到微信支付回调通知")
	ctx := context.Background()
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(constant.WxPayKeyPath)
	if err != nil {
		log.Fatal("load merchant private key error")
	}
	// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
	err = downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, constant.WxPaySerialNo, constant.WxPayMchId, constant.WxPayApiV3Key)
	if err != nil {
		log.Fatal("load merchant cert error")
	}
	// 2. 获取商户号对应的微信支付平台证书访问器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(constant.WxPayMchId)
	// 3. 使用证书访问器初始化 `notify.Handler`
	handler, err := notify.NewRSANotifyHandler(constant.WxPayApiV3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	if err != nil {
		log.Fatal("init notify handler error")
	}

	//wechatPayCert, err := utils.LoadCertificate(constant.WxPayCertPath)
	//// 2. 使用本地管理的微信支付平台证书获取微信支付平台证书访问器
	//certificateVisitor := core.NewCertificateMapWithList([]*x509.Certificate{wechatPayCert})
	//// 3. 使用apiv3 key、证书访问器初始化 `notify.Handler`
	//handler, err := notify.NewRSANotifyHandler(constant.WxPayApiV3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))

	transaction := new(payments.Transaction)
	_, err = handler.ParseNotifyRequest(ctx, c.Request, transaction)
	// 如果验签未通过，或者解密失败
	if err != nil {
		log.Printf("notify handler error: %v", err)
	}
	// 处理通知内容
	//fmt.Println(notifyReq.Summary)
	//fmt.Println(transaction.TransactionId)
	// 在这里处理你的业务逻辑，如更新订单状态
	//log.Printf("Transaction ID: %s", transaction.TransactionId)
	tradeNo := *transaction.OutTradeNo
	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)
	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		log.Printf("查询不到对应订单: %s", tradeNo)
		// 响应微信服务器
		c.JSON(200, gin.H{"code": "ERROR", "message": "找不到订单！"})
		return
	}

	payAmout := *transaction.Amount.Total
	amount := int64(topUp.Amount * 100)
	if payAmout != amount {
		// 响应微信服务器
		log.Printf("待支付金额与支付金额不一致: %v, %v", payAmout, amount)
		c.JSON(200, gin.H{"code": "ERROR", "message": "支付金额错误！"})
		return
	}

	if topUp.Status == "pending" {
		topUp.Status = "success"
		err := topUp.Update()
		if err != nil {
			log.Printf("微信支付回调更新订单失败: %v", topUp)
		}
		//user, _ := model.GetUserById(topUp.UserId, false)
		//user.Quota += topUp.Amount * 500000
		err = model.IncreaseUserQuota(topUp.UserId, topUp.Amount*int(common.QuotaPerUnit))
		if err != nil {
			log.Printf("微信支付回调更新用户失败: %v", topUp)
		}
		log.Printf("微信支付回调更新用户成功 %v", topUp)
		model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用在线充值成功，充值金额: %v，支付金额：%f", common.LogQuota(topUp.Amount*int(common.QuotaPerUnit)), topUp.Money))

		// 更新redis里的状态
		common.RedisSet("topup_order::"+tradeNo, topUp.Status, time.Duration(6)*time.Minute)
	}

	// 响应微信服务器
	c.JSON(200, gin.H{"code": "SUCCESS", "message": "OK"})
}

func WxPayCheckOrder(c *gin.Context) {
	orderId := c.Query("orderId")
	queryFromWx := c.Query("queryFromWx")
	status, error := common.RedisGet("topup_order::" + orderId)
	if error != nil {
		c.JSON(200, gin.H{"message": "failed"})
	}
	if status == "success" || queryFromWx != "1" {
		c.JSON(200, gin.H{"message": "success", "data": status})
	} else {
		// 从微信查询订单状态
		ctx := context.Background()
		client := GetWxPayClient(ctx)
		if client == nil {
			c.JSON(200, gin.H{"message": "error", "data": "当前管理员未配置支付信息"})
			return
		}

		svc := native.NativeApiService{Client: client}
		// 发送请求
		resp, _, err := svc.QueryOrderByOutTradeNo(ctx,
			native.QueryOrderByOutTradeNoRequest{
				Mchid:      core.String(constant.WxPayMchId),
				OutTradeNo: core.String(orderId),
			},
		)
		if err != nil {
			c.JSON(200, gin.H{"message": "error", "data": "查询微信支付订单失败！"})
			return
		}

		if *resp.TradeState == "SUCCESS" {
			LockOrder(orderId)
			defer UnlockOrder(orderId)
			topUp := model.GetTopUpByTradeNo(orderId)
			if topUp == nil {
				log.Printf("查询不到对应订单: %s", orderId)
				// 响应微信服务器
				c.JSON(200, gin.H{"code": "ERROR", "message": "找不到订单！"})
				return
			}

			if topUp.Status == "success" {
				c.JSON(200, gin.H{"message": "success", "data": "success"})
				return
			}

			payAmout := *resp.Amount.Total
			amount := int64(topUp.Amount * 100)
			if payAmout != amount {
				log.Printf("待支付金额与支付金额不一致: %v, %v", payAmout, amount)
				c.JSON(200, gin.H{"code": "ERROR", "message": "支付金额错误！"})
				return
			}

			topUp.Status = "success"
			err := topUp.Update()
			if err != nil {
				log.Printf("微信支付主动查询订单状态，更新订单失败: %v", topUp)
			}
			//user, _ := model.GetUserById(topUp.UserId, false)
			//user.Quota += topUp.Amount * 500000
			err = model.IncreaseUserQuota(topUp.UserId, topUp.Amount*int(common.QuotaPerUnit))
			if err != nil {
				log.Printf("微信支付主动查询订单状态，更新用户失败: %v", topUp)
			}
			log.Printf("微信支付主动查询订单状态，更新用户成功 %v", topUp)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用在线充值成功，充值金额: %v，支付金额：%f", common.LogQuota(topUp.Amount*int(common.QuotaPerUnit)), topUp.Money))

			c.JSON(200, gin.H{"message": "success", "data": "success"})
			return
		} else {
			c.JSON(200, gin.H{"message": "success", "data": status})
			return
		}
	}
}
