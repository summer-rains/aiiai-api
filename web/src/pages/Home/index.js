import React, { useContext, useEffect, useState } from 'react';
import { Card, Col, Row } from '@douyinfe/semi-ui';
import {API, getSystemName, showError, showNotice, timestamp2string} from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';

const Home = () => {
  const [statusState] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');

  const displayNotice = async () => {
    const res = await API.get('/api/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        const htmlNotice = marked(data);
        showNotice(htmlNotice, true);
        localStorage.setItem('notice', data);
      }
    } else {
      showError(message);
    }
  };

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const getStartTimeString = () => {
    const timestamp = statusState?.status?.start_time;
    return statusState.status ? timestamp2string(timestamp) : '';
  };

  useEffect(() => {
    displayNotice().then();
    displayHomePageContent().then();
  }, []);
  return (
    <>
      {homePageContentLoaded && homePageContent === '' ? (
        <div className={"home-container"}>
          <div className={"head"}>
            <div className={"system-name"}>
              {getSystemName()}
            </div>
            <div className={"slogan"}>
              <span style={{color: "mediumorchid"}}>让AI创造更大价值</span>
            </div>
            <div className={"slogan"}>
              {"更全面的模型，更实惠的价格"}
            </div>
            <div className={"slogan"}>
              <span style={{color: "mediumvioletred"}}>1元兑换1美元</span>
            </div>
            <div className={"project"}>
              <span>推荐项目：</span>
              <a
                href='https://github.com/vual/ChatGPT-Next-Web-Pro'
                target='_blank'
                rel='noreferrer'
              >
                ChatGPT-Next-Web-Pro
              </a>
              ，{' '}
              <a
                href='https://github.com/vual/lobe-chat-pro'
                target='_blank'
                rel='noreferrer'
              >
                Lobe-Chat-Pro
              </a>
            </div>
          </div>
          <div className={"content"}>
            <Card className={"flex-item"} headerStyle={{textAlign: "center", fontSize: 32}} title={"问答"}
                  shadows={"always"}>
              <div className={"flex-content"}>
                <Row>OpenAI</Row>
                <Row>ChatGPTPlus</Row>
                <Row>Claude</Row>
                <Row>Azure</Row>
                <Row>Google</Row>
                <Row>国内模型</Row>
              </div>
            </Card>

            <Card className={"flex-item"} headerStyle={{textAlign: "center"}} title={"绘图"} shadows={"always"}>
              <div className={"flex-content"}>
                <Row>Midjourney</Row>
                <Row>Dall-E</Row>
                <Row>Flux</Row>
                <Row>Kling(快手可灵)</Row>
              </div>
            </Card>

            <Card className={"flex-item"} headerStyle={{textAlign: "center"}} title={"音频"} shadows={"always"}>
              <div className={"flex-content"}>
                <Row>Suno</Row>
              </div>
            </Card>

            <Card className={"flex-item"} headerStyle={{textAlign: "center"}} title={"视频"} shadows={"always"}>
              <div className={"flex-content"}>
                <Row>Kling(快手可灵)</Row>
                <Row>Luma</Row>
                <Row>Runway</Row>
              </div>
            </Card>
          </div>
        </div>
      ) : (
        <>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div
              style={{ fontSize: 'larger' }}
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            ></div>
          )}
        </>
      )}
    </>
  );
};

export default Home;
