import React from 'react';
import styles from './index.module.css';

const PageDecorator = () => {
  return (
    <div className={styles.decorator}>
      <div className={styles.globalDecor1}></div>
      <div className={styles.globalDecor2}></div>
      <div className={styles.globalDecor3}></div>
      <div className={styles.globalDecor4}></div>
      <div className={styles.gridBackground}></div>
    </div>
  );
};

export default PageDecorator; 