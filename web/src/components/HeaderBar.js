import React, {useContext, useEffect, useMemo, useState} from 'react';
import {Link, useNavigate} from 'react-router-dom';
import {UserContext} from '../context/User';
import {useSetTheme, useTheme} from '../context/Theme';

import {API, getLogo, getSystemName, isAdmin, isMobile, showSuccess} from '../helpers';
import '../index.css';

import fireworks from 'react-fireworks';

import {
  IconCalendarClock,
  IconChecklistStroked,
  IconComment,
  IconCreditCard,
  IconGift,
  IconHelpCircle,
  IconHistogram,
  IconHomeStroked,
  IconImage,
  IconKey,
  IconLayers,
  IconPriceTag,
  IconSetting,
  IconUser
} from '@douyinfe/semi-icons';
import {Avatar, Dropdown, Layout, Nav, Switch} from '@douyinfe/semi-ui';
import {stringToColor} from '../helpers/render';
import {routerMap} from "./SiderBar.js";

// HeaderBar Buttons
let headerButtons = [
  {
    text: '关于',
    itemKey: 'about',
    to: '/about',
    icon: <IconHelpCircle />,
  },
];

let buttons = [
  {
    text: '首页',
    itemKey: 'home',
    to: '/',
    icon: <IconHomeStroked />,
  },
  // {
  //   text: '测试',
  //   itemKey: 'playground',
  //   to: '/playground',
  //   icon: <IconCommentStroked />,
  // },
];

if (localStorage.getItem('chat_link')) {
  headerButtons.splice(1, 0, {
    name: '聊天',
    to: '/chat',
    icon: 'comments',
  });
}

const HeaderBar = () => {
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();

  const [selectedKeys, setSelectedKeys] = useState(['home']);
  const [showSidebar, setShowSidebar] = useState(false);
  const systemName = getSystemName();
  const logo = getLogo();
  const currentDate = new Date();
  // enable fireworks on new year(1.1 and 2.9-2.24)
  const isNewYear =
    (currentDate.getMonth() === 0 && currentDate.getDate() === 1) ||
    (currentDate.getMonth() === 1 &&
      currentDate.getDate() >= 9 &&
      currentDate.getDate() <= 24);

  async function logout() {
    setShowSidebar(false);
    await API.get('/api/user/logout');
    showSuccess('注销成功!');
    userDispatch({ type: 'logout' });
    localStorage.removeItem('user');
    navigate('/login');
  }

  const handleNewYearClick = () => {
    fireworks.init('root', {});
    fireworks.start();
    setTimeout(() => {
      fireworks.stop();
      setTimeout(() => {
        window.location.reload();
      }, 10000);
    }, 3000);
  };

  const theme = useTheme();
  const setTheme = useSetTheme();

  useEffect(() => {
    if (theme === 'dark') {
      document.body.setAttribute('theme-mode', 'dark');
    }

    if (isNewYear) {
      console.log('Happy New Year!');
    }
  }, []);

  const navButtons = useMemo(
    () => [
      {
        text: '首页',
        itemKey: 'home',
        to: '/',
        icon: <IconHomeStroked />,
      },
      // {
      //   text: '测试',
      //   itemKey: 'playground',
      //   to: '/playground',
      //   icon: <IconCommentStroked />,
      // },
      {
        text: '聊天',
        itemKey: 'chat',
        to: '/chat',
        icon: <IconComment />,
        className: localStorage.getItem('chat_link')
          ? 'semi-navigation-item-normal'
          : 'tableHiddle',
      },
      {
        text: '价格',
        itemKey: 'pricing',
        to: '/pricing',
        icon: <IconPriceTag />,
      },
      {
        text: '渠道',
        itemKey: 'channel',
        to: '/channel',
        icon: <IconLayers />,
        className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle',
      },
      {
        text: '令牌',
        itemKey: 'token',
        to: '/token',
        icon: <IconKey />,
      },
      {
        text: '兑换',
        itemKey: 'redemption',
        to: '/redemption',
        icon: <IconGift />,
        className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle',
      },
      {
        text: '钱包',
        itemKey: 'topup',
        to: '/topup',
        icon: <IconCreditCard />,
      },
      {
        text: '用户',
        itemKey: 'user',
        to: '/user',
        icon: <IconUser />,
        className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle',
      },
      {
        text: '日志',
        itemKey: 'log',
        to: '/log',
        icon: <IconHistogram />,
      },
      {
        text: '数据',
        itemKey: 'detail',
        to: '/detail',
        icon: <IconCalendarClock />,
        className:
          localStorage.getItem('enable_data_export') === 'true'
            ? 'semi-navigation-item-normal'
            : 'tableHiddle',
      },
      {
        text: '绘图',
        itemKey: 'midjourney',
        to: '/midjourney',
        icon: <IconImage />,
        className:
          localStorage.getItem('enable_drawing') === 'true'
            ? 'semi-navigation-item-normal'
            : 'tableHiddle',
      },
      {
        text: '任务',
        itemKey: 'task',
        to: '/task',
        icon: <IconChecklistStroked />,
        className:
          localStorage.getItem('enable_task') === 'true'
            ? 'semi-navigation-item-normal'
            : 'tableHiddle',
      },
      {
        text: '设置',
        itemKey: 'setting',
        to: '/setting',
        icon: <IconSetting />,
      },
    ],
    [
      localStorage.getItem('enable_data_export'),
      localStorage.getItem('enable_drawing'),
      localStorage.getItem('enable_task'),
      localStorage.getItem('chat_link'),
      isAdmin(),
    ],
  );
  useEffect(() => {
    const currentPath = window.location.pathname;
    const matchingButton = navButtons.find(button => button.to === currentPath);
    if (matchingButton) {
      setSelectedKeys([matchingButton.itemKey]);
    } else {
      // If no exact match, check for partial matches (e.g., /midjourney/123 should match /midjourney)
      const partialMatch = navButtons.find(button => currentPath.startsWith(button.to));
      if (partialMatch) {
        setSelectedKeys([partialMatch.itemKey]);
      }
    }
  }, []);

  return (
    <>
      <Layout>
        <div style={{ width: '100%' }}>
          <Nav
            mode={'horizontal'}
            // bodyStyle={{ height: 100 }}
            renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
              const routers = {
                about: '/about',
                login: '/login',
                register: '/register',
                home: '/',
                ...routerMap
              };
              return (
                <Link
                  style={{ textDecoration: 'none' }}
                  to={routers[props.itemKey]}
                >
                  {itemElement}
                </Link>
              );
            }}
            selectedKeys={selectedKeys}
            // items={headerButtons}
            onSelect={(key) => {
              setSelectedKeys([key.itemKey]);
            }}
            header={isMobile()?{
              logo: (
                <img src={logo} alt='logo' style={{ marginRight: '0.75em' }} />
              ),
            }:{
              logo: (
                <img src={logo} alt='logo' />
              ),
              text: systemName,
              style: ({color: "blue"})
            }}
            items={navButtons}
            footer={
              <>
                {isNewYear && (
                  // happy new year
                  <Dropdown
                    position='bottomRight'
                    render={
                      <Dropdown.Menu>
                        <Dropdown.Item onClick={handleNewYearClick}>
                          Happy New Year!!!
                        </Dropdown.Item>
                      </Dropdown.Menu>
                    }
                  >
                    <Nav.Item itemKey={'new-year'} text={'🏮'} style={{marginRight: 10}} />
                  </Dropdown>
                )}
                <Nav.Item itemKey={'about'} icon={<IconHelpCircle />}  style={{marginRight: 10}} />
                <>
                {!isMobile() && (
                    <Switch
                      checkedText='🌞'
                      size={'large'}
                      checked={theme === 'dark'}
                      uncheckedText='🌙'
                      onChange={(checked) => {
                        setTheme(checked);
                      }}
                      style={{marginRight: 10}}
                    />
                  )}
                </>
                {userState.user ? (
                  <>
                    <Dropdown
                      position='bottomRight'
                      render={
                        <Dropdown.Menu>
                          <Dropdown.Item onClick={logout}>退出</Dropdown.Item>
                        </Dropdown.Menu>
                      }
                    >
                      <Avatar
                        size='small'
                        color={stringToColor(userState.user.username)}
                        style={{ margin: 4 }}
                      >
                        {userState.user.username[0]}
                      </Avatar>
                      <span>{userState.user.username}</span>
                    </Dropdown>
                  </>
                ) : (
                  <>
                    <Nav.Item
                      itemKey={'login'}
                      text={'登录'}
                      // icon={<IconKey />}
                    />
                    <Nav.Item
                      itemKey={'register'}
                      text={'注册'}
                      icon={<IconUser />}
                    />
                  </>
                )}
              </>
            }
          ></Nav>
        </div>
      </Layout>
    </>
  );
};

export default HeaderBar;
