import { inject } from "src/core/decorators/inject.decorator";
import { PageComponent } from "src/core/decorators/page.decorator";
import { AfterInit, onChangeDetected, onDestroy, onInit } from "src/core/interfaces/lifecycle.interface";
import { BindableProps } from "src/core/interfaces/pageconf.interface";
import { makeid } from "src/core/tools/random.tools";
import { ErrorInfo } from "src/generated/rpc/apiproto/errors_pb";
import { ApiService } from "src/services/api.service";

@PageComponent({
    template: `
        <h1>Verify</h1>
        <div id="messages" bind-innerHTML="messages"></div>
        `
})
export class VerifyPage implements AfterInit, BindableProps {
    bindProps = {
        messages: ""
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

    AfterInit() {
        var url = new URL(window.location.toString());
        var token = url.searchParams.get("token") || "";
        var email = url.searchParams.get("email") || "";
        this.apiService.verify(token, email).then(() => {
            this.logMsg("verify: success")
        }).catch((err: any) => {
            if (err instanceof ErrorInfo) {
                this.logError('verify:' + (err as ErrorInfo).getViolationType().toString())
            } else {
                this.logError('verify:' + err.toString())
            }
        })
    }

    constructor(@inject(ApiService) private apiService: ApiService) {
    }
}