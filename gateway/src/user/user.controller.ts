import { Controller } from '@nestjs/common';
import { GrpcMethod } from '@nestjs/microservices';
import {
  GATEWAY_USER_SERVICE_NAME,
  GatewayUserServiceController,
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';
import { RegisterUserUseCase } from './usecases/register-user.usecase';

@Controller()
export class UserController implements GatewayUserServiceController {
  constructor(private readonly registerUserUseCase: RegisterUserUseCase) {}

  @GrpcMethod(GATEWAY_USER_SERVICE_NAME, 'RegisterUser')
  async registerUser(
    request: RegisterUserRequest,
  ): Promise<RegisterUserResponse> {
    return this.registerUserUseCase.execute(request);
  }
}
