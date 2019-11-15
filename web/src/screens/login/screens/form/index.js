import React from "react";

import styles from "./index.less";

const LoginForm = props => (
  <div className={styles.login}>
    <form method="post" action="/authorize">
      <p>Login with your version control system username and password.</p>
      <input
        placeholder="Username"
        name="username"
        type="text"
        spellCheck="false"
      />
      <input placeholder="Password" name="password" type="password" />
      <input value="Login" type="submit" />
    </form>
  </div>
);

export default LoginForm;
