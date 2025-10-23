import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { ConfigService } from '@nestjs/config';
import { join } from 'path';
import { AuthController } from './auth.controller';
import { LoginUseCase } from './usecases/login.usecase';
import { RotateRefreshTokensUseCase } from './usecases/rotate-refresh-tokens';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: 'AUTH_SERVICE',
        inject: [ConfigService],
        useFactory: (configService: ConfigService) => {
          const serviceUrl = configService.get<string>('services_urls.auth');
          const protosDir = configService.get<string>('PROTOS_DIRECTORY');

          return {
            transport: Transport.GRPC,
            options: {
              url: serviceUrl,
              package: 'auth',
              protoPath: join(protosDir!, 'auth', 'auth.proto'),
            },
          };
        },
      },
    ]),
  ],
  providers: [LoginUseCase, RotateRefreshTokensUseCase],
  controllers: [AuthController],
})
export class AuthModule {}
