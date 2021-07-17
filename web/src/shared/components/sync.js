import React from "react";
import Icon from "./icons/refresh";
import styles from "./sync.less";

export const Message = () => {
  return (
    <div className={styles.root}>
      <div className={styles.alert}>
        <div>
          <Icon />
        </div>
        <div>Account synchronization in progress</div>
      </div>
    </div>
  );
};
