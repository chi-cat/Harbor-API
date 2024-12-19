import React from 'react';
import { VChart } from "@visactor/react-vchart";
import { useTranslation } from 'react-i18next';
import { renderNumber } from '../../helpers/render';

const TokenCacheChart = ({ data }) => {
  const { t } = useTranslation();

  const spec = {
    type: 'bar',
    data: [{
      id: 'tokenCacheData',
      values: data
    }],
    xField: 'Time',
    yField: 'Value',
    seriesField: 'Type',
    stack: true,
    legends: {
      visible: true,
      selectMode: 'single',
      position: 'top',
      flipPage: true
    },
    title: {
      visible: true,
      text: t('token缓存命中分布'),
      subtext: '',
      alignTo: 'left'
    },
    bar: {
      state: {
        hover: {
          stroke: '#000',
          lineWidth: 1,
        },
      },
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['Type'],
            value: (datum) => renderNumber(datum['Value']),
          },
        ],
      },
    },
    color: {
      specified: {
        'Non-Cache Tokens': '#1890ff',
        'Cache Hit Tokens': '#52c41a'
      },
    },
    padding: [40, 20, 60, 60],
    axis: {
      y: {
        title: {
          text: t('Token数'),
          visible: true,
        },
      },
      x: {
        title: {
          text: t('时间'),
          visible: true,
        },
      },
    },
  };

  return (
    <div style={{ height: 500 }}>
      <VChart spec={spec} option={{ mode: "desktop-browser" }} />
    </div>
  );
};

export default TokenCacheChart; 