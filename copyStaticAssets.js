var shell = require('shelljs');

shell.cp("-R", "static", "dist/");
shell.cp("-R", "templates", "dist/");
shell.cp("-R", "utils/ssr.js", "dist/utils/");