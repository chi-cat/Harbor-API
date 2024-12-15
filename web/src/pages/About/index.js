import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';
import { Layout, Typography, Space, Card, Divider } from '@douyinfe/semi-ui';
import styles from './index.module.css';

const { Text, Title, Paragraph } = Typography;

const About = () => {
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout('');
    }
    setAboutLoaded(true);
  };

  useEffect(() => {
    displayAbout();
  }, []);

  return (
    <>
      {aboutLoaded && about === '' ? (
        <div className={styles.container}>
          <Card className={styles.aboutCard}>
            <Title heading={2} className={styles.mainTitle}>关于我们</Title>
            
            <div className={styles.section}>
              <Title heading={4}>项目介绍</Title>
              <Paragraph>
                NextDreamAPI 是一个强大的 AI API 集成平台，为用户提供便捷的人工智能服务接入方案。
                我们致力于为用户提供最优质的 AI 服务体验，支持多种主流模型，确保稳定可靠的服务质量。
              </Paragraph>
            </div>

            <Divider />
            
            <div className={styles.section}>
              <Title heading={4}>支持我们</Title>
              <Paragraph>
                如果您觉得我们的项目对您有帮助，欢迎通过以下方式支持我们：
              </Paragraph>
              <div className={styles.afdianContainer}>
                <iframe 
                  src="https://afdian.com/leaflet?slug=Nextdream" 
                  width="640" 
                  scrolling="no" 
                  height="200" 
                  frameBorder="0"
                  title="afdian-Nextdream"
                />
              </div>
            </div>

            <Divider />

            <div className={styles.footer}>
              <Space vertical align="center">
                <Text type="secondary">NextAPI © 2024 Nextdream. All Rights Reserved.</Text>
                <Text type="tertiary">Created by gtxy27</Text>
              </Space>
            </div>
          </Card>
        </div>
      ) : (
        <>
          {about.startsWith('https://') ? (
            <iframe
              src={about}
              style={{ width: '100%', height: '100vh', border: 'none' }}
              title="about-content"
            />
          ) : (
            <div
              style={{ fontSize: 'larger' }}
              dangerouslySetInnerHTML={{ __html: about }}
            ></div>
          )}
        </>
      )}
    </>
  );
};

export default About;
