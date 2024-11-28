import { inject } from "src/core/decorators/inject.decorator";
import { PageComponent } from "src/core/decorators/page.decorator";
import { AfterInit, onChangeDetected, onDestroy, onInit } from "src/core/interfaces/lifecycle.interface";
import { BindableProps } from "src/core/interfaces/pageconf.interface";
import { makeid } from "src/core/tools/random.tools";
import { ErrorInfo } from "src/generated/rpc/apiproto/errors_pb";
import { ApiService } from "src/services/api.service";

@PageComponent({
    template: `
        <h1>Hello From Home</h1>
        <div>
        <input placeholder="email" type="text" id="input-email" bind-bothway-value="email"/>
        <input placeholder="username"  type="text" id="input-username" bind-bothway-value="username"/>
        <input placeholder="password" type="password" id="input-password" bind-bothway-value="password"/>
        <input placeholder="repeat password" type="password" id="input-password-check" bind-bothway-value="passwordCheck"/>
        </div>
        <br>
        <button type="button" event-click="signup($event)">Signup</button>
        <button type="button" event-click="login($event)">Login</button>
        <button type="button" event-click="changeUsername($event)"> changeUsername </button>

        <div id="messages" bind-innerHTML="messages"></div>
        `
})
export class HomePage implements BindableProps, onInit, AfterInit, onDestroy, onChangeDetected {
    bindProps = {
        email: "",
        username: "",
        password: "",
        passwordCheck: "",

        messages: "",
    };

    onInit() {
        console.log('Home onInit');
    }

    AfterInit() {
        console.log('Home AfterInit');
    }

    onDestroy(): void {
        console.log('Home onDestroy');
    }

    onChangeDetected(key: string, value: any, newValue: any) {
        console.log(key, ' is being changed from ', value, ' to ', newValue);
    }

    constructor(@inject(ApiService) private apiService: ApiService) {
    }

    logSuccess(innerHTML: string) {
        this.bindProps.messages = '<span class="success">' + (innerHTML) + '</span>'
    }

    logMsg(innerHTML: string) {
        this.bindProps.messages = '<span class="info">' + (innerHTML) + '</span>'
    }
    logError(innerHTML: string) {
        this.bindProps.messages = '<span class="error">' + (innerHTML) + '</span>'
    }

    signup(e: any) {
        this.logMsg("signing up")
        this.apiService.signup(this.bindProps.username, this.bindProps.email, this.bindProps.password).then(() => {
            this.logMsg("signin: success")
        }).catch((err) => {
            if (err instanceof ErrorInfo) {
                this.logError('signin:' + err.toString())
            } else {
                this.logError('signin:' + err.toString())
            }
        })
    }

    login(e: any) {
        this.logMsg("login in")
        let identifier = this.bindProps.username
        if (this.bindProps.username != "") {
            identifier = this.bindProps.email
        }
        this.apiService.login(this.bindProps.username, this.bindProps.password).then(() => {
            this.logMsg("login: success")
        }).catch((err: any) => {
            if (err instanceof ErrorInfo) {
                this.logError('login:' + (err as ErrorInfo).getViolationType().toString())
            } else {
                this.logError('login:' + err.toString())
            }
        })
    }

    changeUsername(e: any) {
        this.logMsg("Changing username")
        this.apiService.changeUsername(this.bindProps.username).then(() => {
            this.logMsg("changeUsername: success" + this.apiService.getUserInfo()!!)
        }).catch((err: any) => {
            if (err instanceof ErrorInfo) {
                this.logError('changeUsername:' + (err as ErrorInfo).getViolationType().toString())
            } else {
                this.logError('changeUsername:' + err.toString())
            }
        })
    }
}