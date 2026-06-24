/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//! Helpers for carrying Redfish diagnostic payload fields into log sinks.
//!
//! Redfish exposes these fields as part of the event or log-entry schema.
//! The health crate intentionally does not parse CPER or other diagnostic
//! formats. It preserves the payload and related Redfish metadata so each sink,
//! or its downstream consumer, can decide how to handle the opaque body.

use std::borrow::Cow;

use serde::Serialize;

use crate::metrics::MetricLabel;
use crate::sink::DiagnosticLogRecord;

const DIAGNOSTIC_DATA_TYPE_ATTR: &str = "redfish.diagnostic_data.type";
const DIAGNOSTIC_DATA_OEM_TYPE_ATTR: &str = "redfish.diagnostic_data.oem_type";
const DIAGNOSTIC_DATA_ADDITIONAL_URI_ATTR: &str = "redfish.diagnostic_data.additional_uri";
const DIAGNOSTIC_DATA_SIZE_BYTES_ATTR: &str = "redfish.diagnostic_data.size_bytes";
const PARENT_MESSAGE_ID_ATTR: &str = "redfish.parent.message_id";
const PARENT_EVENT_ID_ATTR: &str = "redfish.parent.event_id";
const PARENT_LOG_ENTRY_ID_ATTR: &str = "redfish.parent.log_entry_id";

/// Borrowed diagnostic fields extracted from one Redfish event or log entry.
pub(crate) struct DiagnosticPayload<'a> {
    /// Opaque Redfish `DiagnosticData` payload, commonly base64 text for CPER.
    pub diagnostic_data: Option<&'a str>,

    /// Redfish `DiagnosticDataType`, serialized with the schema wire spelling.
    pub diagnostic_data_type: Option<String>,

    /// Vendor-specific diagnostic data type when Redfish provides one.
    pub oem_diagnostic_data_type: Option<&'a str>,

    /// Redfish `AdditionalDataURI`; this module forwards the URI but does not
    /// fetch it.
    pub additional_data_uri: Option<&'a str>,

    /// Optional Redfish size metadata for `AdditionalDataURI`.
    pub additional_data_size_bytes: Option<i64>,

    /// Parent Redfish message id for correlating the diagnostic fields.
    pub message_id: Option<&'a str>,

    /// Parent Redfish event id for correlating the diagnostic fields.
    pub event_id: Option<&'a str>,

    /// Parent Redfish log entry id for correlating the diagnostic fields.
    pub log_entry_id: Option<&'a str>,
}

/// Flattens generated Redfish nullable option fields.
///
/// Redfish generated models use `Option<Option<T>>` to distinguish an absent
/// property from an explicit JSON null. Diagnostic export treats both as absent.
pub(crate) fn nullable_ref<T>(value: &Option<Option<T>>) -> Option<&T> {
    value.as_ref().and_then(Option::as_ref)
}

/// Flattens generated Redfish nullable string fields.
///
/// See [`nullable_ref`] for why explicit null and missing values are handled the
/// same way in diagnostic fields.
pub(crate) fn nullable_str(value: &Option<Option<String>>) -> Option<&str> {
    nullable_ref(value).map(String::as_str)
}

/// Serializes generated Redfish enums through serde to preserve wire spelling.
pub(crate) fn redfish_enum_string<T: Serialize>(value: &T) -> Option<String> {
    serde_json::to_string(value)
        .ok()
        .and_then(|value| serde_json::from_str::<String>(&value).ok())
}

/// Builds diagnostic fields from Redfish payload fields.
///
/// URI-only diagnostics are still forwarded because the URI and size metadata
/// may be enough for a downstream collector to fetch or correlate the payload.
pub(crate) fn make_diagnostic_record(
    payload: DiagnosticPayload<'_>,
) -> Option<DiagnosticLogRecord> {
    if payload.diagnostic_data.is_none() && payload.additional_data_uri.is_none() {
        return None;
    }

    // These attributes leave health through generic sinks, so keep Redfish
    // schema fields namespaced instead of competing with sink-level keys.
    let mut attributes: Vec<MetricLabel> = [
        (
            DIAGNOSTIC_DATA_TYPE_ATTR,
            payload.diagnostic_data_type.as_deref(),
        ),
        (
            DIAGNOSTIC_DATA_OEM_TYPE_ATTR,
            payload.oem_diagnostic_data_type,
        ),
        (
            DIAGNOSTIC_DATA_ADDITIONAL_URI_ATTR,
            payload.additional_data_uri,
        ),
        (PARENT_MESSAGE_ID_ATTR, payload.message_id),
        (PARENT_EVENT_ID_ATTR, payload.event_id),
        (PARENT_LOG_ENTRY_ID_ATTR, payload.log_entry_id),
    ]
    .into_iter()
    .filter_map(|(key, value)| value.map(|value| (Cow::Borrowed(key), value.to_string())))
    .collect();

    if let Some(size_bytes) = payload.additional_data_size_bytes {
        attributes.push((
            Cow::Borrowed(DIAGNOSTIC_DATA_SIZE_BYTES_ATTR),
            size_bytes.to_string(),
        ));
    }

    // Keep the diagnostic body opaque. CPER and vendor formats are parsed by
    // downstream consumers that understand their schema.
    Some(DiagnosticLogRecord {
        body: payload.diagnostic_data.unwrap_or_default().to_string(),
        attributes,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[derive(Serialize)]
    enum TestDiagnosticDataType {
        #[serde(rename = "CPERSection")]
        CperSection,
    }

    fn attr_value<'a>(record: &'a DiagnosticLogRecord, key: &str) -> Option<&'a str> {
        record
            .attributes
            .iter()
            .find(|(k, _)| k.as_ref() == key)
            .map(|(_, v)| v.as_str())
    }

    #[test]
    fn redfish_enum_string_uses_serde_wire_spelling() {
        assert_eq!(
            redfish_enum_string(&TestDiagnosticDataType::CperSection),
            Some("CPERSection".to_string())
        );
    }

    #[test]
    fn diagnostic_payload_fields_preserve_opaque_payload() {
        let Some(record) = make_diagnostic_record(DiagnosticPayload {
            diagnostic_data: Some("base64-payload"),
            diagnostic_data_type: Some("CPER".to_string()),
            oem_diagnostic_data_type: None,
            additional_data_uri: Some("/redfish/v1/Log/1/data"),
            additional_data_size_bytes: Some(2048),
            message_id: Some("ResourceEvent.1.0.ResourceErrorsDetected"),
            event_id: Some("ev-1"),
            log_entry_id: Some("42"),
        }) else {
            panic!("diagnostic payload should produce diagnostic fields");
        };

        assert_eq!(record.body, "base64-payload");

        assert_eq!(
            attr_value(&record, "redfish.diagnostic_data.type"),
            Some("CPER")
        );

        assert_eq!(
            attr_value(&record, "redfish.diagnostic_data.additional_uri"),
            Some("/redfish/v1/Log/1/data")
        );

        assert_eq!(
            attr_value(&record, "redfish.diagnostic_data.size_bytes"),
            Some("2048")
        );

        assert_eq!(attr_value(&record, "redfish.parent.event_id"), Some("ev-1"));

        assert_eq!(
            attr_value(&record, "redfish.parent.log_entry_id"),
            Some("42")
        );
    }

    #[test]
    fn diagnostic_payload_fields_skip_absent_payload_and_uri() {
        let event = make_diagnostic_record(DiagnosticPayload {
            diagnostic_data: None,
            diagnostic_data_type: Some("CPER".to_string()),
            oem_diagnostic_data_type: None,
            additional_data_uri: None,
            additional_data_size_bytes: None,
            message_id: None,
            event_id: None,
            log_entry_id: None,
        });

        assert!(event.is_none());
    }
}
