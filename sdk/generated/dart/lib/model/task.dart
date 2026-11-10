//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//
// @dart=2.18

// ignore_for_file: unused_element, unused_import
// ignore_for_file: always_put_required_named_parameters_first
// ignore_for_file: constant_identifier_names
// ignore_for_file: lines_longer_than_80_chars

part of openapi.api;

class Task {
  /// Returns a new [Task] instance.
  Task({
    required this.taskId,
    this.taskKey,
    this.projectId,
    required this.title,
    this.description,
    this.type,
    this.priority,
    this.severity,
    this.taskWorkflowStatusId,
    this.dueDate,
    this.revision,
    this.billable,
  });

  String taskId;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? taskKey;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? projectId;

  String title;

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
  String? type;

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
  String? severity;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  String? taskWorkflowStatusId;

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
  int? revision;

  ///
  /// Please note: This property should have been non-nullable! Since the specification file
  /// does not include a default value (using the "default:" property), however, the generated
  /// source code must fall back to having a nullable type.
  /// Consider adding a "default:" property in the specification file to hide this note.
  ///
  bool? billable;

  @override
  bool operator ==(Object other) => identical(this, other) || other is Task &&
    other.taskId == taskId &&
    other.taskKey == taskKey &&
    other.projectId == projectId &&
    other.title == title &&
    other.description == description &&
    other.type == type &&
    other.priority == priority &&
    other.severity == severity &&
    other.taskWorkflowStatusId == taskWorkflowStatusId &&
    other.dueDate == dueDate &&
    other.revision == revision &&
    other.billable == billable;

  @override
  int get hashCode =>
    // ignore: unnecessary_parenthesis
    (taskId.hashCode) +
    (taskKey == null ? 0 : taskKey!.hashCode) +
    (projectId == null ? 0 : projectId!.hashCode) +
    (title.hashCode) +
    (description == null ? 0 : description!.hashCode) +
    (type == null ? 0 : type!.hashCode) +
    (priority == null ? 0 : priority!.hashCode) +
    (severity == null ? 0 : severity!.hashCode) +
    (taskWorkflowStatusId == null ? 0 : taskWorkflowStatusId!.hashCode) +
    (dueDate == null ? 0 : dueDate!.hashCode) +
    (revision == null ? 0 : revision!.hashCode) +
    (billable == null ? 0 : billable!.hashCode);

  @override
  String toString() => 'Task[taskId=$taskId, taskKey=$taskKey, projectId=$projectId, title=$title, description=$description, type=$type, priority=$priority, severity=$severity, taskWorkflowStatusId=$taskWorkflowStatusId, dueDate=$dueDate, revision=$revision, billable=$billable]';

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
      json[r'task_id'] = this.taskId;
    if (this.taskKey != null) {
      json[r'task_key'] = this.taskKey;
    } else {
      json[r'task_key'] = null;
    }
    if (this.projectId != null) {
      json[r'project_id'] = this.projectId;
    } else {
      json[r'project_id'] = null;
    }
      json[r'title'] = this.title;
    if (this.description != null) {
      json[r'description'] = this.description;
    } else {
      json[r'description'] = null;
    }
    if (this.type != null) {
      json[r'type'] = this.type;
    } else {
      json[r'type'] = null;
    }
    if (this.priority != null) {
      json[r'priority'] = this.priority;
    } else {
      json[r'priority'] = null;
    }
    if (this.severity != null) {
      json[r'severity'] = this.severity;
    } else {
      json[r'severity'] = null;
    }
    if (this.taskWorkflowStatusId != null) {
      json[r'task_workflow_status_id'] = this.taskWorkflowStatusId;
    } else {
      json[r'task_workflow_status_id'] = null;
    }
    if (this.dueDate != null) {
      json[r'due_date'] = this.dueDate;
    } else {
      json[r'due_date'] = null;
    }
    if (this.revision != null) {
      json[r'revision'] = this.revision;
    } else {
      json[r'revision'] = null;
    }
    if (this.billable != null) {
      json[r'billable'] = this.billable;
    } else {
      json[r'billable'] = null;
    }
    return json;
  }

  /// Returns a new [Task] instance and imports its values from
  /// [value] if it's a [Map], null otherwise.
  // ignore: prefer_constructors_over_static_methods
  static Task? fromJson(dynamic value) {
    if (value is Map) {
      final json = value.cast<String, dynamic>();

      // Ensure that the map contains the required keys.
      // Note 1: the values aren't checked for validity beyond being non-null.
      // Note 2: this code is stripped in release mode!
      assert(() {
        requiredKeys.forEach((key) {
          assert(json.containsKey(key), 'Required key "Task[$key]" is missing from JSON.');
          assert(json[key] != null, 'Required key "Task[$key]" has a null value in JSON.');
        });
        return true;
      }());

      return Task(
        taskId: mapValueOfType<String>(json, r'task_id')!,
        taskKey: mapValueOfType<String>(json, r'task_key'),
        projectId: mapValueOfType<String>(json, r'project_id'),
        title: mapValueOfType<String>(json, r'title')!,
        description: mapValueOfType<String>(json, r'description'),
        type: mapValueOfType<String>(json, r'type'),
        priority: mapValueOfType<String>(json, r'priority'),
        severity: mapValueOfType<String>(json, r'severity'),
        taskWorkflowStatusId: mapValueOfType<String>(json, r'task_workflow_status_id'),
        dueDate: mapValueOfType<String>(json, r'due_date'),
        revision: mapValueOfType<int>(json, r'revision'),
        billable: mapValueOfType<bool>(json, r'billable'),
      );
    }
    return null;
  }

  static List<Task> listFromJson(dynamic json, {bool growable = false,}) {
    final result = <Task>[];
    if (json is List && json.isNotEmpty) {
      for (final row in json) {
        final value = Task.fromJson(row);
        if (value != null) {
          result.add(value);
        }
      }
    }
    return result.toList(growable: growable);
  }

  static Map<String, Task> mapFromJson(dynamic json) {
    final map = <String, Task>{};
    if (json is Map && json.isNotEmpty) {
      json = json.cast<String, dynamic>(); // ignore: parameter_assignments
      for (final entry in json.entries) {
        final value = Task.fromJson(entry.value);
        if (value != null) {
          map[entry.key] = value;
        }
      }
    }
    return map;
  }

  // maps a json object with a list of Task-objects as value to a dart map
  static Map<String, List<Task>> mapListFromJson(dynamic json, {bool growable = false,}) {
    final map = <String, List<Task>>{};
    if (json is Map && json.isNotEmpty) {
      // ignore: parameter_assignments
      json = json.cast<String, dynamic>();
      for (final entry in json.entries) {
        map[entry.key] = Task.listFromJson(entry.value, growable: growable,);
      }
    }
    return map;
  }

  /// The list of required keys that must be present in a JSON.
  static const requiredKeys = <String>{
    'task_id',
    'title',
  };
}
