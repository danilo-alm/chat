import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { UserServiceClient } from 'protos/ts/user/user';
import {
  DeleteUserRequest,
  DeleteUserResponse,
} from 'protos/ts/gateway/gateway';

@Injectable()
export class DeleteUserUseCase implements OnModuleInit {
  private userService: UserServiceClient;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {}

  onModuleInit() {
    this.userService = this.client.getService('UserService');
  }

  async execute({ id }: DeleteUserRequest): Promise<DeleteUserResponse> {
    const observableResponse = this.userService.deleteUser({ id });
    await firstValueFrom(observableResponse).catch((error) => {
      throw new RpcException(error as object);
    });

    return {};
  }
}
