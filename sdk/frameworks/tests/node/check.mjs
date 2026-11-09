import 'reflect-metadata';
import {createServer} from 'node:http';
import assert from 'node:assert/strict';
import express from 'express';
import Fastify from 'fastify';
import {Scope} from '@nestjs/common';
import {REQUEST,ContextIdFactory} from '@nestjs/core';
import {Test} from '@nestjs/testing';
import * as h3 from 'h3';
import {PulseClient} from '../../../typescript/dist/index.js';
import {pulseFromAuthorization,handleTaskList} from '../../../typescript/dist/frameworks.js';
const backend=createServer((req,res)=>{res.setHeader('Content-Type','application/json');res.end(JSON.stringify({tasks:{items:[{title:req.headers.authorization}]}}));});
await new Promise(resolve=>backend.listen(0,'127.0.0.1',resolve));const base=`http://127.0.0.1:${backend.address().port}`;
async function result(user){return handleTaskList(base,new Request('http://localhost/tasks?project_id=demo',{headers:user?{Authorization:user}:{}}));}
const app=express();app.get('/tasks',async(req,res)=>{const r=await result(req.headers.authorization);res.status(r.status).json(await r.json());});
const server=app.listen(0,'127.0.0.1');await new Promise(resolve=>server.once('listening',resolve));
const fast=Fastify();fast.get('/tasks',async(req,reply)=>{const r=await result(req.headers.authorization);return reply.code(r.status).send(await r.json());});await fast.listen({port:0,host:'127.0.0.1'});
const module=await Test.createTestingModule({providers:[{provide:PulseClient,scope:Scope.REQUEST,inject:[REQUEST],useFactory:req=>pulseFromAuthorization(base,req.headers.authorization)}]}).compile();
try {
 for(const port of [server.address().port,fast.server.address().port]){
  assert.equal((await fetch(`http://127.0.0.1:${port}/tasks`)).status,401);
  for(const user of ['alice','bob']){const response=await fetch(`http://127.0.0.1:${port}/tasks`,{headers:{Authorization:`Bearer ${user}`}});assert.equal((await response.json()).tasks.items[0].title,`Bearer ${user}`);}
 }
 for(const user of ['alice','bob']){const ctx=ContextIdFactory.create();module.registerRequestByContextId({headers:{authorization:`Bearer ${user}`}},ctx);const client=await module.resolve(PulseClient,ctx);assert.equal((await client.listTasks('demo')).tasks.items[0].title,`Bearer ${user}`);}
 const nitro=new h3.H3();nitro.get('/tasks',event=>handleTaskList(base,event.req));
 const nitroResponse=await nitro.fetch(new Request('http://localhost/tasks?project_id=demo',{headers:{Authorization:'Bearer nitro'}}));assert.equal((await nitroResponse.json()).tasks.items[0].title,'Bearer nitro');
 console.log(JSON.stringify({h3:'passed',express:'passed',fastify:'passed',nestjs:'passed',h3_toWebRequest:typeof h3.toWebRequest,h3_H3:typeof h3.H3}));
} finally {await module.close();await fast.close();server.closeAllConnections();await new Promise(resolve=>server.close(resolve));backend.closeAllConnections();await new Promise(resolve=>backend.close(resolve));}
