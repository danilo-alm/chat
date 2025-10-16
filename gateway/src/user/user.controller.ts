import { Controller } from '@nestjs/common';
import { GrpcMethod } from '@nestjs/microservices';
import {
  DeleteUserRequest,
  DeleteUserResponse,
  GATEWAY_USER_SERVICE_NAME,
  GatewayUserServiceController,
  GetUserProfileRequest,
  GetUserProfileResponse,
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';
import { RegisterUserUseCase } from './usecases/register-user.usecase';
import { GetUserProfileUseCase } from './usecases/get-user-profile.usecase';
import { DeleteUserUseCase } from './usecases/delete-user.usecase';

@Controller()
export class UserController implements GatewayUserServiceController {
  constructor(
    private readonly registerUserUseCase: RegisterUserUseCase,
    private readonly getUserProfileUseCase: GetUserProfileUseCase,
    private readonly deleteUserUseCase: DeleteUserUseCase,
  ) {}

  @GrpcMethod(GATEWAY_USER_SERVICE_NAME)
  async registerUser(
    request: RegisterUserRequest,
  ): Promise<RegisterUserResponse> {
    return this.registerUserUseCase.execute(request);
  }

  @GrpcMethod(GATEWAY_USER_SERVICE_NAME)
  async getUserProfile(
    request: GetUserProfileRequest,
  ): Promise<GetUserProfileResponse> {
    return this.getUserProfileUseCase.execute(request);
  }

  @GrpcMethod(GATEWAY_USER_SERVICE_NAME)
  deleteUser(request: DeleteUserRequest): Promise<DeleteUserResponse> {
    return this.deleteUserUseCase.execute(request);
  }
}
