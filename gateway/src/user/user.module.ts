import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { ConfigService } from '@nestjs/config';
import { join } from 'path';
import { RegisterUserUseCase } from './usecases/register-user.usecase';
import { UserController } from './user.controller';
import { GetUserProfileUseCase } from './usecases/get-user-profile.usecase';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: 'USER_SERVICE',
        inject: [ConfigService],
        useFactory: (configService: ConfigService) => {
          const serviceUrl = configService.get<string>('services_urls.user');
          const protosDir = configService.get<string>('PROTOS_DIRECTORY');

          return {
            transport: Transport.GRPC,
            options: {
              url: serviceUrl,
              package: 'user',
              protoPath: join(protosDir!, 'user', 'user.proto'),
            },
          };
        },
      },
    ]),
  ],
  providers: [RegisterUserUseCase, GetUserProfileUseCase],
  controllers: [UserController],
})
export class UserModule {}
