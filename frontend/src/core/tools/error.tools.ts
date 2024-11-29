import * as grpcWeb from 'grpc-web';
import { toByteArray } from 'base64-js';

import { pad } from 'src/core/tools/pad.tools'

import { ErrorInfo, ErrorType } from 'src/generated/rpc/apiproto/errors_pb'
import { Status } from 'src/generated/rpc/google_rpc/status_pb'


export const defaultError = new ErrorInfo().setType(ErrorType.ERROR_UNSPECIFIED);

export function retrieveRPCError(err: grpcWeb.RpcError): ErrorInfo {
    if (!("grpc-status-details-bin" in err["metadata"] && typeof err.metadata["grpc-status-details-bin"] === "string")) {
        return defaultError;
    }
    let bytes: Uint8Array;

    try {
       bytes = toByteArray(pad(err.metadata["grpc-status-details-bin"]));
    } catch {
        return defaultError;
    }
    const st = Status.deserializeBinary(bytes);
    const details = st.getDetailsList().map((details: any) => parseErrorDetails(details))
    .filter((details: any): details is ErrorInfo => !!details);
    if (details.length > 0) {
        return details[0];
    }
    return defaultError;
}

function parseErrorDetails(details: any): ErrorInfo | null {
    const typeUrl = details.getTypeUrl();
    if (typeUrl != 'type.googleapis.com/xxxyourappyyy.apiproto.ErrorInfo') {
        return null;
    }
    return ErrorInfo.deserializeBinary(details.getValue_asU8());
}


export function ErrorTypeToString(a : ErrorType) {
    console.log(a.toString());
}