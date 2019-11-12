import React from "react";
import { Route, Switch } from "react-router-dom";
import Title from "react-title-component";

// @see https://github.com/yannickcr/eslint-plugin-react/issues/512
// eslint-disable-next-line react/display-name
export default function() {
	return (
		<Switch>
			<Route path="/account/tokens" exact={true} component={accountTitle} />
			<Route path="/account/repos" exact={true} component={accountRepos} />
			<Route path="/login" exact={false} component={loginTitle} />
			<Route path="/:owner/:repo" exact={false} component={repoTitle} />
			<Route path="/" exact={false} component={defautTitle} />
		</Switch>
	);
}

const accountTitle = () => <Title render="Tokens | drone" />;

const accountRepos = () => <Title render="Repositories | drone" />;

const loginTitle = () => <Title render="Login | drone" />;

const repoTitle = ({ match }) => (
	<Title render={`${match.params.owner}/${match.params.repo} | drone`} />
);

const defautTitle = () => <Title render="Welcome | drone" />;
