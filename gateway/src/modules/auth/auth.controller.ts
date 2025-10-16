import { Controller } from '@nestjs/common';
import {
  GATEWAY_AUTH_SERVICE_NAME,
  GatewayAuthServiceController,
  LoginRequest,
  LoginResponse,
} from 'protos/ts/gateway/gateway';
import { LoginUseCase } from './usecases/login.usecase';
import { GrpcMethod } from '@nestjs/microservices';

@Controller()
export class AuthController implements GatewayAuthServiceController {
  constructor(private readonly loginUseCase: LoginUseCase) {}

  @GrpcMethod(GATEWAY_AUTH_SERVICE_NAME)
  login(request: LoginRequest): Promise<LoginResponse> {
    return this.loginUseCase.execute(request);
  }
}
