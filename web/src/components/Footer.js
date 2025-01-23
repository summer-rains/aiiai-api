import React, { useEffect, useState } from 'react';

import { getFooterHTML, getSystemName } from '../helpers';
import { Layout, Tooltip } from '@douyinfe/semi-ui';

const FooterBar = () => {
  const systemName = getSystemName();
  const [footer, setFooter] = useState(getFooterHTML());
  let remainCheckTimes = 5;

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  const defaultFooter = (
    <div className='custom-footer'>
      <a>
        AIIAI API {import.meta.env.VITE_REACT_APP_VERSION}{' '}
      </a>
      ，{' '}基于{' '}
      <a
        href='https://github.com/Calcium-Ion/new-api'
        target='_blank'
        rel='noreferrer'
      >
        New API
      </a>{' '}
      ，{' '}基于{' '}
      <a
        href='https://github.com/songquanpeng/one-api'
        target='_blank'
        rel='noreferrer'
      >
        One API
      </a>
      <span>，微信：822784588。</span>
      <span>声明：本站api仅限于测试和体验，请自觉遵守法律法规，切勿用于非法用途，本站不承担任何法律责任。</span>
    </div>
  );

  useEffect(() => {
    const timer = setInterval(() => {
      if (remainCheckTimes <= 0) {
        clearInterval(timer);
        return;
      }
      remainCheckTimes--;
      loadFooter();
    }, 200);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div style={{ textAlign: 'center' }}>
      {footer ? (
        <div
          className='custom-footer'
          dangerouslySetInnerHTML={{ __html: footer }}
        ></div>
      ) : (
        defaultFooter
      )}
    </div>
  );
};

export default FooterBar;
