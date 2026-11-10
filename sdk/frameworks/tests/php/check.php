<?php
require __DIR__.'/vendor/autoload.php';
use Riders\Pulse\Api\TasksApi;
use Riders\Pulse\Integration\ClientFactory;
use Riders\Pulse\Laravel\PulseServiceProvider;
$app=new Illuminate\Foundation\Application(__DIR__);
$app->instance('config',new Illuminate\Config\Repository(['pulse'=>['url'=>'https://pulse.example.test','timeout'=>60]]));
$app->instance('request',Illuminate\Http\Request::create('/tasks','GET',[],[],[],['HTTP_AUTHORIZATION'=>'Bearer alice']));
$app->register(PulseServiceProvider::class);
$alice=$app->make(TasksApi::class)->getConfig()->getAccessToken();
$app->forgetScopedInstances();
$app->instance('request',Illuminate\Http\Request::create('/tasks','GET',[],[],[],['HTTP_AUTHORIZATION'=>'Bearer bob']));
$bob=$app->make(TasksApi::class)->getConfig()->getAccessToken();
if($alice!=='alice'||$bob!=='bob')throw new Exception('Laravel identity mixed');
$app->forgetScopedInstances();$app->instance('request',Illuminate\Http\Request::create('/tasks'));
$blocked=false;try{$app->make(ClientFactory::class);}catch(Illuminate\Auth\AuthenticationException){$blocked=true;}
if(!$blocked)throw new Exception('Laravel anonymous accepted');
$container=new Symfony\Component\DependencyInjection\ContainerBuilder();
$stack=new Symfony\Component\HttpFoundation\RequestStack();
$container->register('request_stack',get_class($stack))->setSynthetic(true)->setPublic(true);
$container->setParameter('env(PULSE_URL)','https://pulse.example.test');
$loader=new Symfony\Component\DependencyInjection\Loader\YamlFileLoader($container,new Symfony\Component\Config\FileLocator(__DIR__.'/vendor/riders-agency/pulse-sdk'));
$loader->load('symfony-services.yaml');$container->getDefinition(TasksApi::class)->setPublic(true);$container->compile(true);$container->set('request_stack',$stack);
foreach(['alice','bob'] as $user){$stack->push(Symfony\Component\HttpFoundation\Request::create('/tasks','GET',[],[],[],['HTTP_AUTHORIZATION'=>'Bearer '.$user]));$token=$container->get(TasksApi::class)->getConfig()->getAccessToken();if($token!==$user)throw new Exception('Symfony identity mixed');$stack->pop();}
$blocked=false;try{$container->get(TasksApi::class);}catch(Symfony\Component\HttpKernel\Exception\UnauthorizedHttpException){$blocked=true;}
if(!$blocked)throw new Exception('Symfony anonymous accepted');
echo json_encode(['laravel'=>Illuminate\Foundation\Application::VERSION,'symfony'=>Symfony\Component\HttpKernel\Kernel::VERSION,'request_identity'=>'passed','anonymous_access'=>'denied']),PHP_EOL;
