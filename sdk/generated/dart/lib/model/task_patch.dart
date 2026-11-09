//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class TaskPatch {
  /// Returns a new [TaskPatch] instance.
  TaskPatch({
    this.title,
    this.description,
    this.priority,
    this.dueDate,
    this.billable,
    this.expectedRevision,
  });

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? title;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? description;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? priority;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? dueDate;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  bool? billable;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  int? expectedRevision;

  @override
  bool operator ==(Object other) => identical(this, other) || other is TaskPatch &&
    other.title == title &&
    other.description == description &&
    other.priority == priority &&
    other.dueDate == dueDate &&
    other.billable == billable &&
    other.expectedRevision == expectedRevision;

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (title == null ? 0 : title!.hashCode) +
    (description == null ? 0 : description!.hashCode) +
    (priority == null ? 0 : priority!.hashCode) +
    (dueDate == null ? 0 : dueDate!.hashCode) +
    (billable == null ? 0 : billable!.hashCode) +
    (expectedRevision == null ? 0 : expectedRevision!.hashCode);

  @override
  String toString() => 'TaskPatch[title=$title, description=$description, priority=$priority, dueDate=$dueDate, billable=$billable, expectedRevision=$expectedRevision]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
    if (this.title != null) {
      json[r'title'] = this.title;
    } else {
      json[r'title'] = null;
    }
    if (this.description != null) {
      json[r'description'] = this.description;
    } else {
      json[r'description'] = null;
    }
    if (this.priority != null) {
      json[r'priority'] = this.priority;
    } else {
      json[r'priority'] = null;
    }
    if (this.dueDate != null) {
      json[r'due_date'] = this.dueDate;
    } else {
      json[r'due_date'] = null;
    }
    if (this.billable != null) {
      json[r'billable'] = this.billable;
    } else {
      json[r'billable'] = null;
    }
    if (this.expectedRevision != null) {
      json[r'expected_revision'] = this.expectedRevision;
    } else {
      json[r'expected_revision'] = null;
    }
    return json;
  }

  /// Returns a new [TaskPatch] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static TaskPatch? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "TaskPatch[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "TaskPatch[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return TaskPatch(
        title: mapValueOfType<String>(json, r'title'),
        description: mapValueOfType<String>(json, r'description'),
        priority: mapValueOfType<String>(json, r'priority'),
        dueDate: mapValueOfType<String>(json, r'due_date'),
        billable: mapValueOfType<bool>(json, r'billable'),
        expectedRevision: mapValueOfType<int>(json, r'expected_revision'),
      );
    }
    return null;
  }

  static List<TaskPatch> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <TaskPatch>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = TaskPatch.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, TaskPatch> mapFromJson(dynamic json) {
    final map = <String, TaskPatch>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = TaskPatch.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of TaskPatch-objects as value to a dart map
  static Map<String, List<TaskPatch>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<TaskPatch>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = TaskPatch.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
  };
}
