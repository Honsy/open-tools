var log4js = require('log4js');
log4js.configure({
    appenders: {
        //错误日志 type:过滤类型logLevelFilter,将过滤error日志写进指定文件
        errorLog: { type: 'file', filename: 'logs/error.log' },
        error: { type: "logLevelFilter", level: "error", appender: 'errorLog' },
        infoLog: { type: 'file', filename: 'logs/info.log' },
        info: { type: "logLevelFilter", level: "info", appender: 'infoLog' },
        console: { type: 'console' }
    },
    categories: {
        info: { appenders: ['info'], level: 'info' },
        error: { appenders: ['error'], level: 'error' },
        default: { appenders: ['console', 'error'], level: 'trace' }
    }
});

var errlogger = log4js.getLogger('error');
var infologger = log4js.getLogger('info');

module.exports = {
    errlogger,
    infologger
}