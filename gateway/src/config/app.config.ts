import { join } from 'path';

export default () => ({
  PROTOS_DIRECTORY: join(__dirname, '..', '..', 'protos'),
  services_urls: {
    user: process.env.USER_SERVICE_URL,
    auth: process.env.AUTH_SERVICE_URL,
    last_seen: process.env.LAST_SEEN_URL,
  },
});
