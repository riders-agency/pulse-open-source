//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;


class AIApi {
  AIApi([ApiClient? apiClient]) : apiClient = apiClient ?? defaultApiClient;

  final ApiClient apiClient;

  /// approveProposal
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] toolCallId (required):
  Future<Response> approveProposalWithHttpInfo(String projectId, String toolCallId,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/approve'
      .replaceAll('{project_id}', projectId)
      .replaceAll('{tool_call_id}', toolCallId);

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

  /// approveProposal
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] toolCallId (required):
  Future<Session?> approveProposal(String projectId, String toolCallId,) async {
    final response = await approveProposalWithHttpInfo(projectId, toolCallId,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }

  /// cancelSession
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] sessionId (required):
  Future<Response> cancelSessionWithHttpInfo(String projectId, String sessionId,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/cancel'
      .replaceAll('{project_id}', projectId)
      .replaceAll('{session_id}', sessionId);

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

  /// cancelSession
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] sessionId (required):
  Future<Session?> cancelSession(String projectId, String sessionId,) async {
    final response = await cancelSessionWithHttpInfo(projectId, sessionId,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }

  /// createSession
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [SessionInput] sessionInput (required):
  Future<Response> createSessionWithHttpInfo(String projectId, SessionInput sessionInput,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/employee/sessions'
      .replaceAll('{project_id}', projectId);

    // ignore: prefer_final_locals
    Object? postBody = sessionInput;

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

  /// createSession
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [SessionInput] sessionInput (required):
  Future<Session?> createSession(String projectId, SessionInput sessionInput,) async {
    final response = await createSessionWithHttpInfo(projectId, sessionInput,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }

  /// editProposal
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] toolCallId (required):
  ///
  /// * [ProposalPatch] proposalPatch (required):
  Future<Response> editProposalWithHttpInfo(String projectId, String toolCallId, ProposalPatch proposalPatch,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}'
      .replaceAll('{project_id}', projectId)
      .replaceAll('{tool_call_id}', toolCallId);

    // ignore: prefer_final_locals
    Object? postBody = proposalPatch;

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

  /// editProposal
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] toolCallId (required):
  ///
  /// * [ProposalPatch] proposalPatch (required):
  Future<Session?> editProposal(String projectId, String toolCallId, ProposalPatch proposalPatch,) async {
    final response = await editProposalWithHttpInfo(projectId, toolCallId, proposalPatch,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }

  /// getSession
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] sessionId (required):
  Future<Response> getSessionWithHttpInfo(String projectId, String sessionId,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/employee/sessions/{session_id}'
      .replaceAll('{project_id}', projectId)
      .replaceAll('{session_id}', sessionId);

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

  /// getSession
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] sessionId (required):
  Future<Session?> getSession(String projectId, String sessionId,) async {
    final response = await getSessionWithHttpInfo(projectId, sessionId,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }

  /// rejectProposal
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] toolCallId (required):
  ///
  /// * [Rejection] rejection (required):
  Future<Response> rejectProposalWithHttpInfo(String projectId, String toolCallId, Rejection rejection,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/tool-calls/{tool_call_id}/reject'
      .replaceAll('{project_id}', projectId)
      .replaceAll('{tool_call_id}', toolCallId);

    // ignore: prefer_final_locals
    Object? postBody = rejection;

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

  /// rejectProposal
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] toolCallId (required):
  ///
  /// * [Rejection] rejection (required):
  Future<Session?> rejectProposal(String projectId, String toolCallId, Rejection rejection,) async {
    final response = await rejectProposalWithHttpInfo(projectId, toolCallId, rejection,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }

  /// sendMessage
  ///
  /// Note: This method returns the HTTP [Response].
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] sessionId (required):
  ///
  /// * [MessageInput] messageInput (required):
  Future<Response> sendMessageWithHttpInfo(String projectId, String sessionId, MessageInput messageInput,) async {
    // ignore: prefer_const_declarations
    final path = r'/api/v1/projects/{project_id}/ai/employee/sessions/{session_id}/messages'
      .replaceAll('{project_id}', projectId)
      .replaceAll('{session_id}', sessionId);

    // ignore: prefer_final_locals
    Object? postBody = messageInput;

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

  /// sendMessage
  ///
  /// Parameters:
  ///
  /// * [String] projectId (required):
  ///
  /// * [String] sessionId (required):
  ///
  /// * [MessageInput] messageInput (required):
  Future<Session?> sendMessage(String projectId, String sessionId, MessageInput messageInput,) async {
    final response = await sendMessageWithHttpInfo(projectId, sessionId, messageInput,);
    if (response.statusCode >= HttpStatus.badRequest) {
      throw ApiException(response.statusCode, await _decodeBodyBytes(response));
    }
    // When a remote server returns no body with a status of 204, we shall not decode it.
    // At the time of writing this, `dart:convert` will throw an "Unexpected end of input"
    // FormatException when trying to decode an empty string.
    if (response.body.isNotEmpty && response.statusCode != HttpStatus.noContent) {
      return await apiClient.deserializeAsync(await _decodeBodyBytes(response), 'Session',) as Session;

    }
    return null;
  }
}
