import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { join } from 'path';
import { RegisterUserUseCase } from './usecases/register-user.usecase';

export const USER_INJECTION_TOKEN = 'USER_MICROSERVICE';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: USER_INJECTION_TOKEN,
        imports: [ConfigModule],
        inject: [ConfigService],

        useFactory: (configService: ConfigService) => {
          const serviceUrl = configService.get<string>('services_urls.user');
          const protosDir = configService.get<string>('PROTOS_DIRECTORY');

          return {
            transport: Transport.GRPC,
            options: {
              url: serviceUrl,
              package: 'user',
              protoPath: join(protosDir!, 'user.proto'),
            },
          };
        },
      },
    ]),
  ],
  providers: [RegisterUserUseCase],
  exports: [RegisterUserUseCase],
})
export class UserModule {}
