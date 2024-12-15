import React, { useContext, useEffect, useState } from 'react';
import { Card, Col, Row, Button, Typography, Space, Table, Tag } from '@douyinfe/semi-ui';
import { IconCloud, IconCode, IconSafe, IconShield, IconTick } from '@douyinfe/semi-icons';
import { API, showError, showNotice } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { UserContext } from '../../context/User';
import { marked } from 'marked';
import { StyleContext } from '../../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import styles from './index.module.css';

const { Text, Title, Paragraph } = Typography;

const Home = () => {
  const { t } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [userState] = useContext(UserContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [styleState, styleDispatch] = useContext(StyleContext);
  const navigate = useNavigate();
  const isLoggedIn = userState?.user !== null;

  const handleStartClick = () => {
    if (isLoggedIn) {
      styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
      if (!styleState.isMobile) {
        styleDispatch({ type: 'SET_SIDER', payload: true });
      }
      navigate('/token');  // 已登录用户跳转到令牌页面
    } else {
      navigate('/register');  // 未登录用户跳转到注册页面
    }
  };

  const handleLearnMoreClick = () => {
    const modelsSection = document.querySelector('#models-section');
    if (modelsSection) {
      modelsSection.scrollIntoView({ behavior: 'smooth' });
    }
  };

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
      setHomePageContent('');
    }
    setHomePageContentLoaded(true);
  };

  useEffect(() => {
    displayNotice();
    displayHomePageContent();
  }, []);

  const features = [
    {
      icon: <IconCloud size="extra-large" />,
      title: '多平台集成',
      description: '集成阿里通义千问、腾讯混元、智谱GLM4、百度文心等多家国内领先AI平台，一个账号畅享所有模型'
    },
    {
      icon: <IconCode size="extra-large" />,
      title: '保证原价',
      description: '所有模型均采用官方原价计费，无任何加价，已实现Deepseek的原价计费支持，支持对公付款、支付宝、微信支付'
    },
    {
      icon: <IconSafe size="extra-large" />,
      title: '稳定可靠',
      description: '国内BGP多线服务器部署，已完成备案，访问稳定快速，支持私有部署，提供7×24小时技术支持'
    },
    {
      icon: <IconShield size="extra-large" />,
      title: '安全合规',
      description: '数据安全防护，支持签署保密协议，可提供合规证明，满足企业需求，支持私有化部署'
    }
  ];

  const modelFeatures = [
    {
      category: '通用对话',
      models: [
        { name: '通义千问4.0', features: ['200K上下文', '中文优化', '知识全面'] },
        { name: 'GLM4', features: ['128K上下文', '多模态', '代码能力强'] },
        { name: '月之暗面-KIMI', features: ['32K上下文', '数学推理', '成本低'] },
        { name: 'Deepseek-Chat', features: ['32K上下文', '代码能力强', '性价比高'] }
      ]
    },
    {
      category: '理解与创作',
      models: [
        { name: '文心一言4.0', features: ['知识全面', '中文创作', '多模态'] },
        { name: '腾讯混元', features: ['对话流畅', '任务全面', '成本可控'] },
        { name: '讯飞星火3.0', features: ['中文优化', '垂直领域', '成本低'] },
        { name: 'MiniMax', features: ['对话流畅', '成本可控', '响应快'] }
      ]
    }
  ];

  const pricingPlans = [
    {
      title: '按量计费',
      features: [
        '所有模型官方原价',
        '模型实时更新',
        '价格与官方同步',
        '无最低消费',
        '余额永久有效'
      ]
    },
    {
      title: '企业定制',
      features: [
        '专属技术支持',
        '优先响应保障',
        '专属对接人员',
        '可签保密协议',
        '私有化部署'
      ]
    }
  ];

  // 更新 AI 用例数据
  const aiUseCases = [
    {
      title: '视频创作',
      icon: '🎬',
      description: '一键生成视频脚本、字幕、配音和背景音乐，快速制作短视频和营销内容',
      scenarios: ['视频脚本', '字幕生成', '配音制作', '音乐生成']
    },
    {
      title: '文案创作',
      icon: '✍️',
      description: '快速生成营销文案、产品描述、文章内容，提升内容创作效率',
      scenarios: ['营销文案', '产品文案', '新闻稿件', '社媒内容']
    },
    {
      title: '代码开发',
      icon: '👨‍💻',
      description: '协助程序开发，代码优化建议，bug分析修复，提升开发效率',
      scenarios: ['代码生成', '代码优化', '问题诊断', '技术方案']
    },
    {
      title: '设计创作',
      icon: '🎨',
      description: 'AI辅助生成Logo、海报、UI界面，为设计师提供灵感和创意参考',
      scenarios: ['Logo设计', 'UI设计', '海报设计', '品牌设计']
    },
    {
      title: '文档处理',
      icon: '📄',
      description: '智能处理PDF、Word等文档，支持文本提取、翻译、总结和内容分析',
      scenarios: ['文档总结', '内容分析', '文本翻译', '格式转换']
    },
    {
      title: '搜索总结',
      icon: '🔍',
      description: '智能搜索和信息提取，快速总结长文内容，生成研究报告和竞品分析',
      scenarios: ['网页总结', '竞品分析', '市场调研', '研究报告']
    }
  ];

  return (
    <div className={styles.container}>
      {homePageContentLoaded && homePageContent === '' ? (
        <>
          <div className={styles.hero}>
            <Title heading={1} className={styles.heroTitle}>
              国内领先的AI API集成平台
            </Title>
            <Text className={styles.heroSubtitle}>
              一站式接入多家国内AI服务，稳定可靠，原价计费
            </Text>
            <Space vertical align="center">
              <Button type="primary" theme="solid" size="large" onClick={handleStartClick}>
                {isLoggedIn ? '立即体验' : '免费注册'}
              </Button>
              <Button type="tertiary" theme="solid" size="large" onClick={handleLearnMoreClick}>
                了解更多
              </Button>
            </Space>
          </div>

          <div className={styles.features}>
            <Title heading={2} className={styles.sectionTitle}>
              平台优势
            </Title>
            <Row gutter={[24, 24]}>
              {features.map((feature, index) => (
                <Col xs={24} sm={12} md={6} key={index}>
                  <Card className={styles.featureCard}>
                    <div className={styles.featureIcon}>{feature.icon}</div>
                    <Title heading={4}>{feature.title}</Title>
                    <Text>{feature.description}</Text>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>

          <div id="models-section" className={styles.models}>
            <Title heading={2} className={styles.sectionTitle}>
              全面的模型支持
            </Title>
            <div className={styles.modelCategories}>
              {modelFeatures.map((category, index) => (
                <div key={index} className={styles.modelCategory}>
                  <Title heading={3} className={styles.categoryTitle}>
                    {category.category}
                  </Title>
                  <Row gutter={[16, 16]}>
                    {category.models.map((model, mIndex) => (
                      <Col xs={24} sm={12} md={6} key={mIndex}>
                        <Card className={styles.modelCard}>
                          <Title heading={5}>{model.name}</Title>
                          <Space wrap>
                            {model.features.map((feature, fIndex) => (
                              <Tag key={fIndex} color="blue" size="small">
                                {feature}
                              </Tag>
                            ))}
                          </Space>
                        </Card>
                      </Col>
                    ))}
                  </Row>
                </div>
              ))}
            </div>
          </div>

          {/* AI 用例部分 */}
          <div className={styles.useCases}>
            <Title heading={2} className={styles.sectionTitle}>
              AI 应用场景
            </Title>
            <div className={styles.useCasesContainer}>
              <Row gutter={[24, 24]} justify="center">
                {aiUseCases.map((useCase, index) => (
                  <Col xs={24} sm={12} md={6} key={index}>
                    <Card className={styles.useCaseCard}>
                      <div className={styles.useCaseIcon}>{useCase.icon}</div>
                      <Title heading={4}>{useCase.title}</Title>
                      <Paragraph>{useCase.description}</Paragraph>
                      <div className={styles.scenarios}>
                        {useCase.scenarios.map((scenario, sIndex) => (
                          <Tag key={sIndex} color="blue" size="small">
                            {scenario}
                          </Tag>
                        ))}
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            </div>
          </div>

          <div className={styles.pricing}>
            <Title heading={2} className={styles.sectionTitle}>
              灵活的计费方式
            </Title>
            <Row gutter={[24, 24]}>
              {pricingPlans.map((plan, index) => (
                <Col xs={24} sm={12} key={index}>
                  <Card className={styles.pricingCard}>
                    <Title heading={3}>{plan.title}</Title>
                    <ul className={styles.pricingFeatures}>
                      {plan.features.map((feature, fIndex) => (
                        <li key={fIndex}>
                          <IconTick className={styles.tickIcon} />
                          <Text>{feature}</Text>
                        </li>
                      ))}
                    </ul>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>

          <div className={styles.cta}>
            <Title heading={2}>开始使用</Title>
            <Text className={styles.ctaText}>
              {isLoggedIn ? '立即体验国内领先的AI服务' : '立即注册账号，体验国内领先的AI服务'}
            </Text>
            <Space vertical align="center">
              <Button type="primary" theme="solid" size="large" onClick={handleStartClick}>
                {isLoggedIn ? '立即体验' : '免费注册'}
              </Button>
              {!isLoggedIn && (
                <Text className={styles.ctaSubtext}>
                  已有账号？ <a onClick={() => navigate('/login')}>立即登录</a>
                </Text>
              )}
            </Space>
          </div>
        </>
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
    </div>
  );
};

export default Home;
