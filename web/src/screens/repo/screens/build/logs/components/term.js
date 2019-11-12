import React, { Component } from "react";
import AnsiUp from "ansi_up";
import style from "./term.less";
import { Link } from "react-router-dom";

let formatter = new AnsiUp();
formatter.use_classes = true;

class Term extends Component {
	render() {
		const { lines, exitcode, highlighted } = this.props;
		return (
			<div className={style.term}>
				{lines.map(line => renderTermLine(line, highlighted))}
				{exitcode !== undefined ? renderExitCode(exitcode) : undefined}
			</div>
		);
	}

	shouldComponentUpdate(nextProps, nextState) {
		return (
			this.props.lines !== nextProps.lines ||
			this.props.exitcode !== nextProps.exitcode ||
			this.props.highlighted !== nextProps.highlighted
		);
	}
}

class TermLine extends Component {
	render() {
		const { line, highlighted } = this.props;
		return (
			<div
				className={highlighted === line.pos ? style.highlight : style.line}
				key={line.pos}
				ref={highlighted === line.pos ? ref => (this.ref = ref) : null}
			>
				<div>
					<Link to={`#L${line.pos + 1}`} key={line.pos + 1}>
						{line.pos + 1}
					</Link>
				</div>
				<div dangerouslySetInnerHTML={{ __html: this.colored }} />
				<div>{line.time || 0}s</div>
			</div>
		);
	}

	componentDidMount() {
		if (this.ref !== undefined) {
			scrollToRef(this.ref);
		}
	}

	get colored() {
		return formatter.ansi_to_html(this.props.line.out || "");
	}

	shouldComponentUpdate(nextProps, nextState) {
		return (
			this.props.line.out !== nextProps.line.out ||
			this.props.highlighted !== nextProps.highlighted
		);
	}
}

const renderTermLine = (line, highlighted) => {
	return <TermLine line={line} highlighted={highlighted} />;
};

const renderExitCode = code => {
	return <div className={style.exitcode}>exit code {code}</div>;
};

const TermError = () => {
	return (
		<div className={style.error}>
			Oops. There was a problem loading the logs.
		</div>
	);
};

const TermLoading = () => {
	return <div className={style.loading}>Loading ...</div>;
};

const scrollToRef = ref => window.scrollTo(0, ref.offsetTop - 100);

Term.Line = TermLine;
Term.Error = TermError;
Term.Loading = TermLoading;

export default Term;
