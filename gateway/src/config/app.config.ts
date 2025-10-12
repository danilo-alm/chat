import { resolve } from 'path';

export default () => {
  const config = {
    PROTOS_DIRECTORY: resolve(__dirname, '..', '..', '..', '..', 'protos'),
    services_urls: {
      user: process.env.USER_SERVICE_URL,
      auth: process.env.AUTH_SERVICE_URL,
      last_seen: process.env.LAST_SEEN_URL,
    },
  };

  // const required = ['USER_SERVICE_URL', 'AUTH_SERVICE_URL', 'LAST_SEEN_URL'];
  const required = [];
  const missing = required.filter((key) => !process.env[key]);

  if (missing.length > 0) {
    throw new Error(
      `Missing required environment variables: ${missing.join(', ')}`,
    );
  }
  return config;
};
