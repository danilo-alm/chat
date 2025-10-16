import { Inject, Injectable } from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { CreateUserRequest, UserServiceClient } from 'protos/ts/user/user';
import {
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';

@Injectable()
export class RegisterUserUseCase {
  private userService: UserServiceClient;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {
    this.userService = this.client.getService<UserServiceClient>('UserService');
  }

  async execute(req: RegisterUserRequest): Promise<RegisterUserResponse> {
    const serviceRequest: CreateUserRequest = {
      name: req.name,
      username: req.username,
      password: req.password,
    };

    const observableResponse = this.userService.createUser(serviceRequest);
    const serviceResponse = await firstValueFrom(observableResponse).catch(
      (error) => {
        throw new RpcException(error as object);
      },
    );
    return {
      id: serviceResponse.id,
    };
  }
}
