import React from 'react';
import { VChart } from "@visactor/react-vchart";
import { useTranslation } from 'react-i18next';
import { renderQuota, renderQuotaNumberWithDigit } from '../../helpers/render';

const ConsumptionChart = ({ data, modelColors, consumeQuota }) => {
  const { t } = useTranslation();

  const spec = {
    type: 'bar',
    data: [{
      id: 'barData',
      values: data
    }],
    xField: 'Time',
    yField: 'Usage',
    seriesField: 'Model',
    stack: true,
    legends: {
      visible: true,
      selectMode: 'single',
      position: 'top',
      flipPage: true
    },
    title: {
      visible: true,
      text: t('模型消耗分布'),
      subtext: `${t('总计')}：${renderQuota(consumeQuota, 2)}`,
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
            key: (datum) => datum['Model'],
            value: (datum) =>
              renderQuotaNumberWithDigit(parseFloat(datum['Usage']), 4),
          },
        ],
      },
      dimension: {
        content: [
          {
            key: (datum) => datum['Model'],
            value: (datum) => datum['Usage'],
          },
        ],
        updateContent: (array) => {
          array.sort((a, b) => b.value - a.value);
          let sum = 0;
          for (let i = 0; i < array.length; i++) {
            sum += parseFloat(array[i].value);
            array[i].value = renderQuotaNumberWithDigit(
              parseFloat(array[i].value),
              4,
            );
          }
          array.unshift({
            key: t('总计'),
            value: renderQuotaNumberWithDigit(sum, 4),
          });
          return array;
        },
      },
    },
    color: {
      specified: modelColors,
    },
    padding: [40, 20, 60, 60],
    axis: {
      y: {
        title: {
          text: t('消耗量'),
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

export default ConsumptionChart; 