"use strict";
exports.__esModule = true;
exports.getContent = exports.getRawHeader = exports.getHeader = void 0;
var marked_1 = require("marked");
var tokens = ['---', '---'];
var regexHeader = new RegExp('^' + tokens[0] + '([\\s|\\S]*?)' + tokens[1]);
var regexContent = new RegExp('^ *?\\' + tokens[0] + '[^]*?' + tokens[1] + '*');
function getHeader(data) {
    var header = getRawHeader(data);
    var tmpObj = {};
    var lines = header.trim().split('\n');
    lines.forEach(function (line, i) {
        var arr = line.trim().split(':');
        tmpObj[arr.shift()] = arr.join(':').trim();
    });
    return tmpObj;
}
exports.getHeader = getHeader;
function getRawHeader(data) {
    var header = regexHeader.exec(data);
    if (!header) {
        new Error("Can't get the header");
    }
    return header[1];
}
exports.getRawHeader = getRawHeader;
function getContent(data) {
    var content = data.replace(regexContent, '').replace(/<!--(.*?)-->/gm, '');
    if (!content) {
        throw new Error("Can't get the content");
    }
    return (0, marked_1["default"])(content);
}
exports.getContent = getContent;
