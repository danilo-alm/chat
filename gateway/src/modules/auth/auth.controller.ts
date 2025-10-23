import { Controller } from '@nestjs/common';
import {
  GATEWAY_AUTH_SERVICE_NAME,
  GatewayAuthServiceController,
  LoginRequest,
  LoginResponse,
  RotateRefreshTokenRequest,
  RotateRefreshTokenResponse,
} from 'protos/ts/gateway/gateway';
import { LoginUseCase } from './usecases/login.usecase';
import { GrpcMethod } from '@nestjs/microservices';
import { RotateRefreshTokensUseCase } from './usecases/rotate-refresh-tokens';

@Controller()
export class AuthController implements GatewayAuthServiceController {
  constructor(
    private readonly loginUseCase: LoginUseCase,
    private readonly rotateRefreshTokensUseCase: RotateRefreshTokensUseCase,
  ) {}

  @GrpcMethod(GATEWAY_AUTH_SERVICE_NAME)
  login(request: LoginRequest): Promise<LoginResponse> {
    return this.loginUseCase.execute(request);
  }

  @GrpcMethod(GATEWAY_AUTH_SERVICE_NAME)
  rotateRefreshToken(
    request: RotateRefreshTokenRequest,
  ): Promise<RotateRefreshTokenResponse> {
    return this.rotateRefreshTokensUseCase.execute(request);
  }
}
