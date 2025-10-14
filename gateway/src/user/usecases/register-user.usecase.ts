import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { CreateUserRequest, UserServiceClient } from 'protos/ts/user/user';
import {
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';

@Injectable()
export class RegisterUserUseCase implements OnModuleInit {
  private userService: UserServiceClient;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {}

  onModuleInit() {
    this.userService = this.client.getService<UserServiceClient>('UserService');
  }

  async execute(req: RegisterUserRequest): Promise<RegisterUserResponse> {
    const serviceRequest: CreateUserRequest = {
      name: req.name,
      username: req.username,
      password: req.password,
    };

    try {
      const observableResponse = this.userService.createUser(serviceRequest);
      const serviceResponse = await firstValueFrom(observableResponse);
      return {
        id: serviceResponse.id,
      };
    } catch (error) {
      throw new RpcException(error as object);
    }
  }
}
