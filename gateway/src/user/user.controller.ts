import { Controller } from '@nestjs/common';
import { GrpcMethod } from '@nestjs/microservices';
import {
  GATEWAY_USER_SERVICE_NAME,
  GatewayUserServiceController,
  GetUserProfileRequest,
  GetUserProfileResponse,
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';
import { RegisterUserUseCase } from './usecases/register-user.usecase';
import { GetUserProfileUseCase } from './usecases/get-user-profile.usecase';

@Controller()
export class UserController implements GatewayUserServiceController {
  constructor(
    private readonly registerUserUseCase: RegisterUserUseCase,
    private readonly getUserProfileUseCase: GetUserProfileUseCase,
  ) {}

  @GrpcMethod(GATEWAY_USER_SERVICE_NAME, 'RegisterUser')
  async registerUser(
    request: RegisterUserRequest,
  ): Promise<RegisterUserResponse> {
    return this.registerUserUseCase.execute(request);
  }

  @GrpcMethod(GATEWAY_USER_SERVICE_NAME, 'GetUserProfile')
  async getUserProfile(
    request: GetUserProfileRequest,
  ): Promise<GetUserProfileResponse> {
    return this.getUserProfileUseCase.execute(request);
  }
}
