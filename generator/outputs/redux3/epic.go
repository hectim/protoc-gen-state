package redux3

const EpicTemplate = `/* THIS FILE IS GENERATED FROM THE TOOL PROTOC-GEN-STATE  */
/* ANYTHING YOU EDIT WILL BE OVERWRITTEN IN FUTURE BUILDS */

import { combineEpics } from 'redux-observable';
import { isActionOf } from 'typesafe-actions';
import { Observable } from 'rxjs';
import _ from 'lodash';
import { grpc } from 'grpc-web-client';
import { UnaryOutput } from 'grpc-web-client/dist/unary';
import 'rxjs/add/observable/dom/ajax';
import { toMessage } from './to_message_pb';
import * as protocActions from './actions_pb';
import * as ProtocTypes from './protoc_types_pb';
import * as ProtocServices from './protoc_services_pb';


function noop() {
	return;
}

function createErrorObject(code: number|string|undefined, message: string): NodeJS.ErrnoException {
	var err: NodeJS.ErrnoException = new Error();
	err.message = message;
	if(code && typeof code == 'number') { err.code = code.toString(); }
	if(code && typeof code == 'string') { err.code = code; }
	return err;
}

function createHostString(hostname, hostnameLocation, port, store) {
  let host = ""
  if (hostname != "") {
    host = hostname + port
  } else if (hostnameLocation != "" ) {
    let keys = hostnameLocation.split(".")
    let host = store.getState()[keys[0]]
    for (let i = 1; i < keys.length; i ++) { // only enters this for loop if keys array is larger than 1
      console.log('host: ', host)
      host = host[keys[i]]
    }
    if (host == "" || host == undefined || host == null) {
      throw new Error("PROTOC-GEN-STATE: the value of hostnameLocation is empty. Check that this path is set in redux: "+ hostnameLocation)
    }
    // last char
    if (host.charAt(host.length - 1) == '/') {
      host = host.slice(0,-1) + port
    } else {
      host = host + port
    }
  } else {
    // hostnameLocation and host is empty
    throw new Error("PROTOC-GEN-STATE: hostnameLocation is empty. Check that it's value in redux is set.")
  }
  return host
}

{{range $i, $e := .}}
export const {{$e.Name}}Epic = (action$, store) => action$
	.filter(isActionOf(protocActions.{{$e.Name}}Request))
	.debounceTime({{$e.Debounce}})
	.map(({ payload, meta: { resolve = noop, reject = noop } }) => ({
		message: toMessage(payload, {{$e.ProtoInputType}}),
		resolve,
		reject,
	}))
	.flatMap((request) => {
{{if $e.Repeat}} {{template "grpcStream" $e}} {{ else }} {{template "grpcUnary" $e}} {{end}}
		.retry({{$e.Retries}})
		.timeout({{$e.Timeout}}){{if $e.Updater}}
		.map(obj => ({ ...obj } as { prev: {{$e.ProtoOutputType}}.AsObject, updated: {{$e.ProtoOutputType}}.AsObject } ))
		.map(lib => {
			request.resolve(lib.prev, lib.updated);
			return protocActions.{{$e.Name}}Success(lib);
		}){{else}}
		.map((resObj: {{$e.ProtoOutputType}}.AsObject{{if $e.Repeat}}[]{{end}}) => {
			request.resolve(resObj);
			return protocActions.{{$e.Name}}Success(resObj);
		}){{end}}
		.catch(error => {
			const err: NodeJS.ErrnoException = createErrorObject(error.code, error.message);
			if(request.reject){ request.reject(err); }
			return Observable.of(protocActions.{{$e.Name}}Failure(err));
		})
	})
	.takeUntil(action$.filter(isActionOf(protocActions.{{$e.Name}}Cancel)))
	.repeat();
{{end}}
{{define "grpcUnary"}}   return Observable
		.defer(() => new Promise((resolve, reject) => {
      {{if .Debug}}console.log('calling {{.FullMethodName}} with payload: ', request.message);{{end}}
      var host = createHostString('{{.Hostname}}', '{{.HostnameLocation}}', '{{.Port}}', store)
			{{template "authToken" .}}
			grpc.unary({{.FullMethodName}}, {
				request: request.message,
				host: host,
				{{template "authFollowUp" .}}
				onEnd: (res: UnaryOutput<{{.ProtoOutputType}}>) => {
          {{if .Debug}}console.log('onEnd {{.FullMethodName}}: ', res.message);{{end}}
					if(res.status != grpc.Code.OK){
            {{if .Debug}}console.log('Error in epic -- status: ', res.status, ' message: ', res.statusMessage);{{end}}
						const err: NodeJS.ErrnoException = createErrorObject(res.status, res.statusMessage);
						reject(err);
					}
					if(res.message){
						resolve(res.message.toObject());
					}
				}
			});
		})){{end}}
{{define "grpcStream"}}  var host = createHostString('{{.Hostname}}', '{{.HostnameLocation}}', '{{.Port}}', store)
		return Observable
			.defer(() => new Promise((resolve, reject) => {
        {{if .Debug}}console.log('calling {{.FullMethodName}} with payload: ', request.message);{{end}}
				var arr: {{.ProtoOutputType}}.AsObject[] = [];
				const client = grpc.client({{.FullMethodName}}, {
					host: host,
				});
				client.onMessage((message: {{.ProtoOutputType}}) => {
          {{if .Debug}}console.log('in {{.FullMethodName}} streaming message: ', message.toObject());{{end}}
					arr.push(message.toObject());
				});
        {{if .Debug}}client.onEnd((code: grpc.Code, msg: string, trailers: grpc.Metadata) => {
          console.log('in {{.FullMethodName}} streaming onEnd: ', code, msg, trailers, request.message);{{else}}client.onEnd((code: grpc.Code, msg: string) => { {{end}}
					if (code != grpc.Code.OK) {
            {{if .Debug}}console.log('Error in streaming epic -- code: ', code, ' message: ', msg);{{end}}
						reject(createErrorObject(code, msg));
					}
					resolve(arr);
				});
				client.start({{template "authToken" .}});
				client.send(request.message);
			})){{end}}

export const protocEpics = combineEpics({{range $i, $e := .}}
	{{$e.Name}}Epic,{{end}}
)

{{define "authToken"}} {{if .Auth}} {{if .Repeat}} new grpc.Metadata({ 'Authorization': `+ "`" +`Bearer ${store.getState().{{.Auth}}}` + "`" + ` }) {{else}} var idToken = store.getState().{{.Auth}}; {{end}} {{end}}
{{end}}

{{define "authFollowUp"}} {{if .Auth}} {{if .Repeat}} {{else}} metadata: new grpc.Metadata({ 'Authorization': ` + "`" + `Bearer ${idToken}` + "`" + `}), {{end}} {{end}}
{{end}}
`
