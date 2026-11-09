//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class ToolCall {
  /// Returns a new [ToolCall] instance.
  ToolCall({
    required this.aiToolCallId,
    required this.toolKey,
    required this.status,
    this.requiresApproval,
    this.inputJson = const {},
    this.outputJson = const {},
  });

  String aiToolCallId;

  String toolKey;

  String status;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  bool? requiresApproval;

  Map<String, Object> inputJson;

  Map<String, Object> outputJson;

  @override
  bool operator ==(Object other) => identical(this, other) || other is ToolCall &&
    other.aiToolCallId == aiToolCallId &&
    other.toolKey == toolKey &&
    other.status == status &&
    other.requiresApproval == requiresApproval &&
    _deepEquality.equals(other.inputJson, inputJson) &&
    _deepEquality.equals(other.outputJson, outputJson);

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (aiToolCallId.hashCode) +
    (toolKey.hashCode) +
    (status.hashCode) +
    (requiresApproval == null ? 0 : requiresApproval!.hashCode) +
    (inputJson.hashCode) +
    (outputJson.hashCode);

  @override
  String toString() => 'ToolCall[aiToolCallId=$aiToolCallId, toolKey=$toolKey, status=$status, requiresApproval=$requiresApproval, inputJson=$inputJson, outputJson=$outputJson]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
      json[r'ai_tool_call_id'] = this.aiToolCallId;
      json[r'tool_key'] = this.toolKey;
      json[r'status'] = this.status;
    if (this.requiresApproval != null) {
      json[r'requires_approval'] = this.requiresApproval;
    } else {
      json[r'requires_approval'] = null;
    }
      json[r'input_json'] = this.inputJson;
      json[r'output_json'] = this.outputJson;
    return json;
  }

  /// Returns a new [ToolCall] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static ToolCall? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "ToolCall[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "ToolCall[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return ToolCall(
        aiToolCallId: mapValueOfType<String>(json, r'ai_tool_call_id')!,
        toolKey: mapValueOfType<String>(json, r'tool_key')!,
        status: mapValueOfType<String>(json, r'status')!,
        requiresApproval: mapValueOfType<bool>(json, r'requires_approval'),
        inputJson: mapCastOfType<String, Object>(json, r'input_json') ?? const {},
        outputJson: mapCastOfType<String, Object>(json, r'output_json') ?? const {},
      );
    }
    return null;
  }

  static List<ToolCall> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <ToolCall>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = ToolCall.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, ToolCall> mapFromJson(dynamic json) {
    final map = <String, ToolCall>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = ToolCall.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of ToolCall-objects as value to a dart map
  static Map<String, List<ToolCall>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<ToolCall>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = ToolCall.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
    'ai_tool_call_id',
    'tool_key',
    'status',
  };
}
