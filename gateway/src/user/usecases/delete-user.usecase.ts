import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { DeleteUserRequest, UserServiceClient } from 'protos/ts/user/user';

@Injectable()
export class DeleteUserUseCase implements OnModuleInit {
  private userService: UserServiceClient;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {}

  onModuleInit() {
    this.userService = this.client.getService('UserService');
  }

  async execute(req: DeleteUserRequest): Promise<void> {
    const serviceRequest: DeleteUserRequest = {
      id: req.id,
    };

    const observableResponse = this.userService.deleteUser(serviceRequest);
    await firstValueFrom(observableResponse).catch((error) => {
      throw new RpcException(error as object);
    });
  }
}
