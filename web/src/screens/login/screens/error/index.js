import React, { Component } from "react";
import queryString from "query-string";
import Icon from "shared/components/icons/report";

import styles from "./index.less";

const DEFAULT_ERROR = "The system failed to process your Login request.";

class Error extends Component {
  render() {
    const parsed = queryString.parse(window.location.search);
    let error = DEFAULT_ERROR;

    switch (parsed.code || parsed.error) {
      case "oauth_error":
        break;
      case "access_denied":
        break;
    }

    return (
      <div className={styles.root}>
        <div className={styles.alert}>
          <div>
            <Icon />
          </div>
          <div>{error}</div>
        </div>
      </div>
    );
  }
}

export default Error;
