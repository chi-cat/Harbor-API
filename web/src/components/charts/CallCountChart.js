import React from 'react';
import { VChart } from "@visactor/react-vchart";
import { useTranslation } from 'react-i18next';
import { renderNumber } from '../../helpers/render';

const CallCountChart = ({ data, modelColors, times }) => {
  const { t } = useTranslation();

  const spec = {
    type: 'pie',
    data: [{
      id: 'id0',
      values: data
    }],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    pie: {
      style: {
        cornerRadius: 10,
      },
      state: {
        hover: {
          outerRadius: 0.85,
          stroke: '#000',
          lineWidth: 1,
        },
        selected: {
          outerRadius: 0.85,
          stroke: '#000',
          lineWidth: 1,
        },
      },
    },
    title: {
      visible: true,
      text: t('模型调用次数占比'),
      subtext: `${t('总计')}：${renderNumber(times)}`,
      alignTo: 'left'
    },
    legends: {
      visible: true,
      orient: 'left',
      position: 'top',
      flipPage: true
    },
    label: {
      visible: true,
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['type'],
            value: (datum) => renderNumber(datum['value']),
          },
        ],
      },
    },
    color: {
      specified: modelColors,
    },
    padding: [40, 20, 60, 60],
  };

  return (
    <div style={{ height: 500 }}>
      <VChart spec={spec} option={{ mode: "desktop-browser" }} />
    </div>
  );
};

export default CallCountChart; 