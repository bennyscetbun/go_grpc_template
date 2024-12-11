import { ChangeUsernameReply, ChangeUsernameRequest, LoginReply, LoginRequest, SignupReply, SignupRequest, UserInfo, VerifyEmailReply, VerifyEmailRequest } from "src/generated/rpc/apiproto/api_pb";
import { ApiClient } from "src/generated/rpc/apiproto/ApiServiceClientPb";
import * as grpcWeb from 'grpc-web';
import { defaultError, retrieveRPCError } from "src/core/tools/error.tools";
import { ErrorInfo } from "src/generated/rpc/apiproto/errors_pb";


export class ApiService {
    private client: ApiClient
    private token: string | null
    private userInfo: UserInfo | null

    constructor() {
        this.client = new ApiClient("");
        this.token = null;
        this.userInfo = null;
    }

    getUserInfo(): UserInfo | null {
        return this.userInfo
    }

    login(identifier: string, password: string): Promise<void> {
        let req = new LoginRequest().setIdentifier(identifier).setPassword(password);
        return new Promise<void>((resolve, reject) => {
            this.client.login(req, null).then((value: LoginReply) => {
                this.token = value.getToken()
                if (value.hasUserinfo()) {
                    this.userInfo = value.getUserinfo()!!
                } else {
                    // TODO error
                }
                resolve();
            }, (reason: any) => {
                if (reason instanceof grpcWeb.RpcError) {
                    let errorInfo = retrieveRPCError(reason);
                    reject(errorInfo);
                } else {
                    console.log(reason);
                    reject(defaultError);
                }
            }).catch((err: any) => {
                console.log(err);
                reject(defaultError);
            })
        })
    }

    signup(username: string, email: string, password: string): Promise<void> {
        let req = new SignupRequest().setEmail(email).setUsername(username).setPassword(password);
        return new Promise<void>((resolve, reject) => {
            this.client.signup(req, null).then((value: SignupReply) => {
                this.token = value.getToken()
                if (value.hasUserinfo()) {
                    this.userInfo = value.getUserinfo()!!
                } else {
                    // TODO error
                }
                resolve();
            }, (reason: any) => {
                if (reason instanceof grpcWeb.RpcError) {
                    let errorInfo = retrieveRPCError(reason);
                    reject(errorInfo);
                } else {
                    console.log(reason);
                    reject(defaultError);
                }
            }).catch((err: any) => {
                console.log(err);
                reject(defaultError);
            })
        })
    }

    changeUsername(username: string): Promise<void> {
        let req = new ChangeUsernameRequest().setNewusername(username);
        return new Promise<void>((resolve, reject) => {
            this.client.changeUsername(req, { "authorization": this.token!! }).then((value: ChangeUsernameReply) => {
                if (value.hasUserinfo()) {
                    this.userInfo = value.getUserinfo()!!
                } else {
                    // TODO error
                }
                resolve();
            }, (reason: any) => {
                if (reason instanceof grpcWeb.RpcError) {
                    let errorInfo = retrieveRPCError(reason);
                    reject(errorInfo);
                } else {
                    console.log(reason);
                    reject(defaultError);
                }
            }).catch((err: any) => {
                console.log(err);
                reject(defaultError);
            })
        })
    }

    verify(token: string, email: string):Promise<void> {
        let req = new VerifyEmailRequest().setEmail(email).setVerifyid(token);
        return new Promise<void>((resolve, reject) => {
            this.client.verifyEmail(req, null).then((value: VerifyEmailReply) => {
                if (value.hasUserinfo()) {
                    this.userInfo = value.getUserinfo()!!
                } else {
                    // TODO error
                }
                resolve();
            }, (reason: any) => {
                if (reason instanceof grpcWeb.RpcError) {
                    let errorInfo = retrieveRPCError(reason);
                    reject(errorInfo);
                } else {
                    console.log(reason);
                    reject(defaultError);
                }
            }).catch((err: any) => {
                console.log(err);
                reject(defaultError);
            })
        })
    }
}

