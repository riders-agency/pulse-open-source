import json,threading,importlib.metadata
from http.server import BaseHTTPRequestHandler,ThreadingHTTPServer
from pulse_sdk.frameworks import from_request
class Handler(BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_GET(self):
  body=json.dumps({'tasks':{'items':[{'title':self.headers.get('Authorization')}]}}).encode();self.send_response(200);self.end_headers();self.wfile.write(body)
server=ThreadingHTTPServer(('127.0.0.1',0),Handler);thread=threading.Thread(target=server.serve_forever,daemon=True);thread.start();base=f'http://127.0.0.1:{server.server_port}'
try:
 from fastapi import FastAPI,Request,HTTPException
 from fastapi.testclient import TestClient
 fast=FastAPI()
 @fast.get('/tasks')
 def tasks(request:Request):
  try:return from_request(base,request).list_tasks('project')
  except ValueError:raise HTTPException(status_code=401)
 with TestClient(fast) as client:
  assert client.get('/tasks').status_code==401
  for user in ('alice','bob'):assert client.get('/tasks',headers={'Authorization':f'Bearer {user}'}).json()['tasks']['items'][0]['title']==f'Bearer {user}'
 from flask import Flask,request,jsonify,abort
 flask=Flask(__name__)
 @flask.get('/tasks')
 def flask_tasks():
  try:return jsonify(from_request(base,request).list_tasks('project'))
  except ValueError:abort(401)
 with flask.test_client() as client:
  assert client.get('/tasks').status_code==401
  for user in ('alice','bob'):assert client.get('/tasks',headers={'Authorization':f'Bearer {user}'}).json['tasks']['items'][0]['title']==f'Bearer {user}'
 from django.conf import settings
 settings.configure(ROOT_URLCONF=__name__,ALLOWED_HOSTS=['testserver'],SECRET_KEY='synthetic-local-test-only')
 import django;django.setup()
 from django.http import JsonResponse
 from django.urls import path
 def django_tasks(request):
  try:return JsonResponse(from_request(base,request).list_tasks('project'))
  except ValueError:return JsonResponse({'error':'user_session_required'},status=401)
 urlpatterns=[path('tasks',django_tasks)]
 from django.test import Client
 client=Client();assert client.get('/tasks').status_code==401
 for user in ('alice','bob'):assert client.get('/tasks',HTTP_AUTHORIZATION=f'Bearer {user}').json()['tasks']['items'][0]['title']==f'Bearer {user}'
 print(json.dumps({'frameworks':{name:importlib.metadata.version(name) for name in ('fastapi','django','flask')},'request_identity':'passed','anonymous_access':'denied'}))
finally:server.shutdown();server.server_close();thread.join()
