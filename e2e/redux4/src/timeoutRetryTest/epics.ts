/*
 * The purpose of this file is to test implementation of timouts and retries
 * before adding it to gen-state and needing to rerender and debug c++ code
 * Once it works in here the logic is transferred into protoc-gen-state
 */


import { grpc } from 'grpc-web-client';
import { combineEpics, Epic } from 'redux-observable';

// import { Observable } from 'rxjs/Rx';
// import 'rxjs/add/observable/merge';
// import 'rxjs/add/observable/from';

import { of, Observable, defer, from } from 'rxjs';
import { filter, map, flatMap, debounceTime, tap, catchError, timeout, retry } from 'rxjs/operators';
import 'rxjs/add/observable/dom/ajax';

import { isActionOf } from 'typesafe-actions';

import { RootAction } from '../rootAction';
import { RootState } from '../rootState';
import * as actions from './actions';

import _ from 'lodash';

import * as ProtocServices from 'protos/BasicState/protoc_services_pb';
import * as ProtocTypes from 'protos/BasicState/protoc_types_pb';
import { toMessage } from 'protos/BasicState/to_message_pb';

// const host: string = 'http://35.192.235.78:9090';
// const host: string = 'https://localhost:9091';
// const badhost: string = 'http://www.google.com:81';


const timeoutEpic: Epic<RootAction, RootState> = (action$) => action$.pipe(
  filter(isActionOf(actions.timeoutRequestPromise))
  ,tap((action:any) => { console.log('right here: ', action); })
  ,debounceTime(1000)
  ,flatMap((action:any) => {
    // return Observable
    return 
      // just set a timer longer than the timeout
      from(() => new Promise((resolve) => {
        setTimeout(() =>
          resolve({ title: 'Ulysses', author: 'James Joyce' }),
          3000
        )
      })).pipe(
        // hit this timeout
        timeout(100) // <-- bingo
        // never makes it here but oh well
        ,map(a => {
          console.log('no error: ', a);
          action.payload.resolve(a as ProtocTypes.readinglist.Book.AsObject);
          return actions.timeoutSuccess(a as ProtocTypes.readinglist.Book.AsObject)
        })
        // catch and reject
        ,catchError((err) => {
          console.log('error: ', err);
          action.payload.reject(err);
          return of(actions.timeoutFailure(err));
        })
      )
  })
)

const retryEpic: Epic<RootAction, RootState> = (action$) => action$.pipe(
  filter(isActionOf(actions.retryRequestPromise))
  ,tap((action:any) => { console.log('right here: ', action); })
  ,debounceTime(1000)
  ,flatMap((action:any) => {
    // return Observable
    return
      from(() => new Promise((resolve, reject) => {
        let counter = 0
        setTimeout(() => {
          if(counter > 2) {
            resolve({ title: 'Ulysses', author: 'James Joyce' })
          } else {
            counter += 1
            reject('retry')
          }
        },
          100
        )
      })).pipe(
        map(a => {
          console.log('no error: ', a);
          action.payload.resolve(a as ProtocTypes.readinglist.Book.AsObject);
          return actions.retrySuccess(a as ProtocTypes.readinglist.Book.AsObject)
        })
        ,catchError((err) => {
          // Observable.merge( // <-- come back to this yikes
          //   Observable.of(actions.retryRequestPromise(...action.payload)),
          //   source
          // )
          console.log('error: ', err);
          action.payload.reject(err);
          return of(actions.retryFailure(err));
        })
        ,retry()
      )
  })
)

const codeEpic = (action$, store) => action$.pipe(
  filter(isActionOf(actions.codeRequestPromise))
  ,map((action:any) => ({ ...action.payload, request: toMessage(action.payload.book, ProtocTypes.readinglist.Book)}))
  ,flatMap((action: any) => {
    // return Observable
    return
      from(() => new Promise((resolve, reject) => {
        var host = store.getState()['config']['host'].slice(0, -1) + ":9090";
        grpc.unary(ProtocServices.readinglist.ReadingList.ErrorOut, {
          request: action.request,
          host: host,
          onEnd: (res:any) => {
            console.log("onEnd: ", res);
            if (res.status != grpc.Code.OK) {
              reject(new Error(`grpc-web request failed with status code: ${res.status}`));
            }
            if(res.message) {
              resolve(res.message.toObject());
            }
          }
        });
      })).pipe(
        timeout(3000)
        ,map(resObj => {
          action.resolve(resObj as ProtocTypes.readinglist.Book.AsObject);
          return actions.codeSuccess(resObj as ProtocTypes.readinglist.Book.AsObject);
        })
        ,catchError(error => {
          action.reject(error.toString());
          return of(actions.codeFailure(error.toString()));
        })
      )
  })
)

export const TimeoutRetryEpics = combineEpics(
  timeoutEpic,
  retryEpic,
  codeEpic,
);
