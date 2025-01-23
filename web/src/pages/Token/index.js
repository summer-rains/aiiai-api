import React from 'react';
import TokensTable from '../../components/TokensTable';
import { Banner, Layout } from '@douyinfe/semi-ui';
const Token = () => (
  <>
    <Layout>
      <Layout.Header>
        <Banner
          type='info'
          description={
            <div>
              <div>1.点击下面的<span style={{color: 'var(--semi-color-secondary)'}}>复制</span>按钮，复制对应令牌，即apiKey，填到对应参数设置里</div>
              <div>2.接口地址填：https://ai.aiiai.top，或https://api.annyun.cn，不同应用或模型，可能还要拼上不同的路径，如/v1、/mj、/luma等</div>
              <div>3.<a style={{textDecoration: "underline"}} href='https://gpt-best.apifox.cn' target='_blank' rel='noreferrer'>接口文档</a></div>
            </div>
          }
        />
      </Layout.Header>
      <Layout.Content>
        <TokensTable/>
      </Layout.Content>
    </Layout>
  </>
);

export default Token;
