import winston from 'winston';
import path from 'path';

const logDir = process.env.LOG_FILE_PATH || './logs';

const logFormat = winston.format.combine(
  winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
  winston.format.errors({ stack: true }),
  winston.format.splat(),
  winston.format.json()
);

const winstonLogger = winston.createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: logFormat,
  defaultMeta: { service: 'r2s-backend' },
  transports: [
    new winston.transports.Console({
      format: winston.format.combine(
        winston.format.colorize(),
        winston.format.simple()
      )
    }),
    new winston.transports.File({
      filename: path.join(logDir, 'error.log'),
      level: 'error',
      maxsize: 10485760, // 10MB
      maxFiles: 5
    }),
    new winston.transports.File({
      filename: path.join(logDir, 'combined.log'),
      maxsize: 10485760, // 10MB
      maxFiles: 5
    })
  ]
});

const auditLogger = winston.createLogger({
  level: 'info',
  format: logFormat,
  defaultMeta: { service: 'r2s-audit' },
  transports: [
    new winston.transports.File({
      filename: path.join(logDir, 'audit.log'),
      maxsize: 10485760,
      maxFiles: 10
    })
  ]
});

export const logger = {
  debug: (message: string, ...meta: any[]) => winstonLogger.debug(message, ...meta),
  info: (message: string, ...meta: any[]) => winstonLogger.info(message, ...meta),
  warn: (message: string, ...meta: any[]) => winstonLogger.warn(message, ...meta),
  error: (message: string, ...meta: any[]) => winstonLogger.error(message, ...meta),
  audit: (data: any) => {
    auditLogger.info('AUDIT', data);
  }
};