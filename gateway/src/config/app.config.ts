import { join } from 'path';

export default () => ({
  PROTOS_DIRECTORY: join(__dirname, '..', '..', 'protos'),
  services_urls: {
    user: process.env.USER_SERVICE_URL,
  },
});
