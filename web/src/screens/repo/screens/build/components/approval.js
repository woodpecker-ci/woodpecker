import React from "react";
import style from "./approval.less";

export const Approval = ({ onapprove, ondecline }) => (
	<div className={style.root}>
		<p>Pipeline execution is blocked pending administrator approval</p>
		<button onClick={onapprove}>Approve</button>
		<button onClick={ondecline}>Decline</button>
	</div>
);
