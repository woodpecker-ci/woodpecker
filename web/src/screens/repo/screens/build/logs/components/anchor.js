import React from "react";

import styles from "./anchor.less";

export const Top = () => <div className={styles.top} />;

export const Bottom = () => <div className={styles.bottom} />;

export const scrollToTop = () => {
  document.querySelector(`.${styles.top}`).scrollIntoView();
};

export const scrollToBottom = () => {
  document.querySelector(`.${styles.bottom}`).scrollIntoView();
};
