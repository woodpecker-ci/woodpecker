import React from "react";
import { mount } from "enzyme";

import Status from "../status";
import {
	STATUS_FAILURE,
	STATUS_RUNNING,
	STATUS_SUCCESS,
} from "shared/constants/status";

jest.dontMock("../status");

describe("Status component", () => {
	test("updates on status change", () => {
		const status = mount(<Status status={STATUS_FAILURE} />);
		const instance = status.instance();

		expect(
			instance.shouldComponentUpdate({ status: STATUS_FAILURE }),
		).toBeFalsy();
		expect(
			instance.shouldComponentUpdate({ status: STATUS_SUCCESS }),
		).toBeTruthy();
		expect(status.hasClass("failure")).toBeTruthy();
	});

	test("uses the status as the class name", () => {
		const status = mount(<Status status={STATUS_RUNNING} />);

		expect(status.hasClass("running")).toBeTruthy();
	});
});
