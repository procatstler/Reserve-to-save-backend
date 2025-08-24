import { createClient } from 'redis';
import { logger } from '@utils/logger';

const redisUrl = process.env.REDIS_URL || 'redis://localhost:6379';
const redisPassword = process.env.REDIS_PASSWORD;

export const redis = createClient({
  url: redisUrl,
  password: redisPassword || undefined,
});

redis.on('error', (err) => {
  logger.error('Redis Client Error', err);
});

redis.on('connect', () => {
  logger.info('Redis connected successfully');
});

export async function connectRedis(): Promise<void> {
  await redis.connect();
}

export async function disconnectRedis(): Promise<void> {
  await redis.quit();
  logger.info('Redis connection closed');
}

// Helper functions
export async function setWithExpiry(
  key: string,
  value: string,
  ttlSeconds: number
): Promise<void> {
  await redis.setEx(key, ttlSeconds, value);
}

export async function getAndDelete(key: string): Promise<string | null> {
  const value = await redis.get(key);
  if (value) {
    await redis.del(key);
  }
  return value;
}