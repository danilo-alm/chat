import { Inject, Injectable } from '@nestjs/common';
import { ClientGrpc } from '@nestjs/microservices';
import { CreateUserRequest, UserService } from 'protos/ts/user/user';
import {
  RegisterUserRequest,
  RegisterUserResponse,
} from 'protos/ts/gateway/gateway';
import { lastValueFrom } from 'rxjs';

@Injectable()
export class RegisterUserUseCase {
  private userService: UserService;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {
    this.userService = this.client.getService('UserService');
  }

  async execute(req: RegisterUserRequest): Promise<RegisterUserResponse> {
    const serviceRequest: CreateUserRequest = {
      name: req.name,
      username: req.username,
      password: req.password,
    };
    const createdUser = await lastValueFrom(
      this.userService.CreateUser(serviceRequest),
    );
    return { id: createdUser.id };
  }
}
