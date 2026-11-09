//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;


class TasksApi {
  TasksApi([ApiClient? apiClient]) : apiClient = apiClient ?? defaultApiClient;

  final ApiClient apiClient;

  /// archiveTask
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  Future<Response> archiveTaskWithHttpInfo(String taskId,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/work/tasks/{task_id}'
      .replaceAll('{task_id}', taskId);

    // ignore: prefer_final_locals
    Object? postBody;

    final queryParams = <QueryParam>[];
    final headerParams = <String, String>{};
    final formParams = <String, String>{};

    const contentTypes = <String>[];


    return apiClient.invokeAPI(
      path,
      'DELETE',
      queryParams,
      postBody,
      headerParams,
      formParams,
      contentTypes.isEmpty ? null : contentTypes.first,
    );
  }

  /// archiveTask
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  Future<ArchiveResult?> archiveTask(String taskId,) async {
    final response = await archiveTaskWithHttpInfo(taskId,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'ArchiveResult',) as ArchiveResult;

    }
    return null;
  }

  /// createTask
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [TaskInput] taskInput (required):
  Future<Response> createTaskWithHttpInfo(TaskInput taskInput,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/work/tasks';

    // ignore: prefer_final_locals
    Object? postBody = taskInput;

    final queryParams = <QueryParam>[];
    final headerParams = <String, String>{};
    final formParams = <String, String>{};

    const contentTypes = <String>['application/json'];


    return apiClient.invokeAPI(
      path,
      'POST',
      queryParams,
      postBody,
      headerParams,
      formParams,
      contentTypes.isEmpty ? null : contentTypes.first,
    );
  }

  /// createTask
  ///
  /// Parameters:
  ///
  /// * [TaskInput] taskInput (required):
  Future<Task?> createTask(TaskInput taskInput,) async {
    final response = await createTaskWithHttpInfo(taskInput,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Task',) as Task;

    }
    return null;
  }

  /// getTask
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  Future<Response> getTaskWithHttpInfo(String taskId,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/work/tasks/{task_id}'
      .replaceAll('{task_id}', taskId);

    // ignore: prefer_final_locals
    Object? postBody;

    final queryParams = <QueryParam>[];
    final headerParams = <String, String>{};
    final formParams = <String, String>{};

    const contentTypes = <String>[];


    return apiClient.invokeAPI(
      path,
      'GET',
      queryParams,
      postBody,
      headerParams,
      formParams,
      contentTypes.isEmpty ? null : contentTypes.first,
    );
  }

  /// getTask
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  Future<TaskEnvelope?> getTask(String taskId,) async {
    final response = await getTaskWithHttpInfo(taskId,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'TaskEnvelope',) as TaskEnvelope;

    }
    return null;
  }

  /// listTasks
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [int] limit:
  ///
  /// * [int] offset:
  ///
  /// * [String] search:
  ///
  /// * [bool] archived:
  Future<Response> listTasksWithHttpInfo(String projectId, { int? limit, int? offset, String? search, bool? archived, }) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/work/tasks';

    // ignore: prefer_final_locals
    Object? postBody;

    final queryParams = <QueryParam>[];
    final headerParams = <String, String>{};
    final formParams = <String, String>{};

      queryParams.addAll(_queryParams('', 'project_id', projectId));
    if (limit != null) {
      queryParams.addAll(_queryParams('', 'limit', limit));
    }
    if (offset != null) {
      queryParams.addAll(_queryParams('', 'offset', offset));
    }
    if (search != null) {
      queryParams.addAll(_queryParams('', 'search', search));
    }
    if (archived != null) {
      queryParams.addAll(_queryParams('', 'archived', archived));
    }

    const contentTypes = <String>[];


    return apiClient.invokeAPI(
      path,
      'GET',
      queryParams,
      postBody,
      headerParams,
      formParams,
      contentTypes.isEmpty ? null : contentTypes.first,
    );
  }

  /// listTasks
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [int] limit:
  ///
  /// * [int] offset:
  ///
  /// * [String] search:
  ///
  /// * [bool] archived:
  Future<TaskPage?> listTasks(String projectId, { int? limit, int? offset, String? search, bool? archived, }) async {
    final response = await listTasksWithHttpInfo(projectId,  limit: limit, offset: offset, search: search, archived: archived, );
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'TaskPage',) as TaskPage;

    }
    return null;
  }

  /// restoreTask
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  Future<Response> restoreTaskWithHttpInfo(String taskId,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/work/tasks/{task_id}/restore'
      .replaceAll('{task_id}', taskId);

    // ignore: prefer_final_locals
    Object? postBody;

    final queryParams = <QueryParam>[];
    final headerParams = <String, String>{};
    final formParams = <String, String>{};

    const contentTypes = <String>[];


    return apiClient.invokeAPI(
      path,
      'POST',
      queryParams,
      postBody,
      headerParams,
      formParams,
      contentTypes.isEmpty ? null : contentTypes.first,
    );
  }

  /// restoreTask
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  Future<ArchiveResult?> restoreTask(String taskId,) async {
    final response = await restoreTaskWithHttpInfo(taskId,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'ArchiveResult',) as ArchiveResult;

    }
    return null;
  }

  /// updateTask
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  ///
  /// * [TaskPatch] taskPatch (required):
  Future<Response> updateTaskWithHttpInfo(String taskId, TaskPatch taskPatch,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/work/tasks/{task_id}'
      .replaceAll('{task_id}', taskId);

    // ignore: prefer_final_locals
    Object? postBody = taskPatch;

    final queryParams = <QueryParam>[];
    final headerParams = <String, String>{};
    final formParams = <String, String>{};

    const contentTypes = <String>['application/json'];


    return apiClient.invokeAPI(
      path,
      'PATCH',
      queryParams,
      postBody,
      headerParams,
      formParams,
      contentTypes.isEmpty ? null : contentTypes.first,
    );
  }

  /// updateTask
  ///
  /// Parameters:
  ///
  /// * [String] taskId (required):
  ///
  /// * [TaskPatch] taskPatch (required):
  Future<Task?> updateTask(String taskId, TaskPatch taskPatch,) async {
    final response = await updateTaskWithHttpInfo(taskId, taskPatch,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Task',) as Task;

    }
    return null;
  }
}
