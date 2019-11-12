module.exports = {
	"extends": [
		"standard",
		"plugin:jest/recommended",
		"plugin:react/recommended",
		"prettier",
		"prettier/react"
	],
	"plugins": [
		"react",
		"jest",
		"prettier"
	],
	"parser": "babel-eslint",
	"parserOptions": {
		"ecmaVersion": 2016,
		"sourceType": "module",
		"ecmaFeatures": {
			"jsx": true
		}
	},
	"env": {
		"es6": true,
		"browser": true,
		"node": true,
		"jest/globals": true
	},
	"rules": {
		"react/prop-types": 1,
		"prettier/prettier": [
			"error",
			{
				"trailingComma": "all",
				"useTabs": true
			}
		]
	}
};
