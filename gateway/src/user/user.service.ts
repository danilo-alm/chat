import { Controller } from '@nestjs/common';
import {
  DeleteUserRequest,
  GatewayUserService,
  GetUserProfileRequest,
  GetUserProfileResponse,
  RegisterUserRequest,
  RegisterUserResponse,
  GatewayUserServiceServiceName,
} from 'protos/ts/gateway/gateway';
import { RegisterUserUseCase } from './usecases/register-user.usecase';
import { GetUserProfileUseCase } from './usecases/get-user-profile.usecase';
import { DeleteUserUseCase } from './usecases/delete-user.usecase';
import { Empty } from 'protos/ts/google/protobuf/empty';
import { GrpcMethod } from '@nestjs/microservices';

@Controller()
export class GatewayUserServiceImpl implements GatewayUserService {
  constructor(
    private readonly registerUserUseCase: RegisterUserUseCase,
    private readonly getUserProfileUseCase: GetUserProfileUseCase,
    private readonly deleteUserUseCase: DeleteUserUseCase,
  ) {}

  @GrpcMethod(GatewayUserServiceServiceName)
  RegisterUser(request: RegisterUserRequest): Promise<RegisterUserResponse> {
    return this.registerUserUseCase.execute(request);
  }

  @GrpcMethod(GatewayUserServiceServiceName)
  GetUserProfile(
    request: GetUserProfileRequest,
  ): Promise<GetUserProfileResponse> {
    return this.getUserProfileUseCase.execute(request);
  }

  @GrpcMethod(GatewayUserServiceServiceName)
  DeleteUser(request: DeleteUserRequest): Promise<Empty> {
    return this.deleteUserUseCase.execute(request);
  }
}
