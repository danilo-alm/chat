import { Inject, Injectable } from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { UserServiceClient } from 'protos/ts/user/user';
import {
  DeleteUserRequest,
  DeleteUserResponse,
} from 'protos/ts/gateway/gateway';

@Injectable()
export class DeleteUserUseCase {
  private userService: UserServiceClient;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {
    this.userService = this.client.getService('UserService');
  }

  async execute(req: DeleteUserRequest): Promise<DeleteUserResponse> {
    const observableResponse = this.userService.deleteUser(req);
    await firstValueFrom(observableResponse).catch((error) => {
      throw new RpcException(error as object);
    });
    return {};
  }
}
