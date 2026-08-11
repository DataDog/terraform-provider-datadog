#!/usr/bin/env python3
"""Build-verified tfgen coverage sweep over the Datadog v2 OpenAPI spec."""

from __future__ import annotations

import argparse
import copy
import csv
import dataclasses
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
from typing import Any, Iterable, Optional

try:
    import yaml
except ImportError:
    sys.exit("PyYAML is required")


GET = "get"
GEN_PREFIX = "tfgen_coverage_"
TARGET_PATHS = (
    "datadog/fwprovider",
    "datadog/tests",
    "docs/data-sources",
    "examples/data-sources",
)
HIGH_VALUE = {
    "api_key",
    "application_key",
    "app",
    "dashboard",
    "incident",
    "integration",
    "logs_archive",
    "logs_index",
    "monitor",
    "notebook",
    "on_call",
    "reference_table",
    "restriction_policy",
    "role",
    "security_rule",
    "service_account",
    "slo",
    "spans_metric",
    "team",
    "user",
    "workflow",
}
CSV_FIELDS = (
    "run_date",
    "api_spec_commit",
    "sdk_version",
    "service",
    "resource",
    "resource_path",
    "operation_id",
    "cardinality",
    "endpoint_set",
    "both_mode",
    "annotated",
    "sdk_bound",
    "outcome",
    "failure_class",
    "gap_bucket",
    "gap_stage",
    "gap_confidence",
    "gap_diagnostic",
    "benefit",
    "value_weighted",
)


def eprint(*args: Any) -> None:
    print(*args, file=sys.stderr, flush=True)


def snake(value: str) -> str:
    value = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", value)
    return re.sub(r"[^a-z0-9]+", "_", value.lower()).strip("_")


def singularize(value: str) -> str:
    value = snake(value)
    if value.endswith("ies") and len(value) > 3:
        return value[:-3] + "y"
    if value.endswith(("sses", "shes", "ches", "xes", "zes")):
        return value[:-2]
    if value.endswith("s") and not value.endswith(("ss", "us")):
        return value[:-1]
    return value


def pluralize(value: str) -> str:
    value = singularize(value)
    if value.endswith("y") and not value.endswith(("ay", "ey", "oy", "uy")):
        return value[:-1] + "ies"
    if value.endswith(("s", "x", "z", "ch", "sh")):
        return value + "es"
    return value + "s"


def percent(num: float, den: float) -> str:
    return "n/a" if not den else f"{100.0 * num / den:.1f}%"


def run_command(
    argv: list[str],
    cwd: Path,
    timeout: Optional[int] = None,
    check: bool = False,
) -> subprocess.CompletedProcess[str]:
    proc = subprocess.run(
        argv,
        cwd=cwd,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        timeout=timeout,
    )
    if check and proc.returncode:
        raise RuntimeError(f"{' '.join(argv)} failed ({proc.returncode}):\n{proc.stdout}")
    return proc


class SpecNavigator:
    def __init__(self, spec: dict[str, Any]):
        self.spec = spec

    def resolve_ref(self, node: Any, seen: Optional[set[str]] = None) -> Any:
        seen = set() if seen is None else seen
        while isinstance(node, dict) and "$ref" in node:
            ref = node["$ref"]
            if not isinstance(ref, str) or ref in seen:
                return {}
            seen.add(ref)
            cur: Any = self.spec
            for part in ref.lstrip("#/").split("/"):
                part = part.replace("~1", "/").replace("~0", "~")
                if not isinstance(cur, dict) or part not in cur:
                    return {}
                cur = cur[part]
            node = cur
        return node or {}

    def expanded(self, node: Any, seen: Optional[set[int]] = None) -> dict[str, Any]:
        """Resolve refs/allOf enough to inspect response envelope structure."""
        seen = set() if seen is None else seen
        node = self.resolve_ref(node)
        if not isinstance(node, dict) or id(node) in seen:
            return {}
        seen.add(id(node))
        merged: dict[str, Any] = {}
        for part in node.get("allOf") or []:
            merged = self._merge_schema(merged, self.expanded(part, seen.copy()))
        own = {k: v for k, v in node.items() if k not in ("$ref", "allOf")}
        return self._merge_schema(merged, own)

    @staticmethod
    def _merge_schema(left: dict[str, Any], right: dict[str, Any]) -> dict[str, Any]:
        out = dict(left)
        for key, value in right.items():
            if key == "properties" and isinstance(value, dict):
                props = dict(out.get("properties") or {})
                props.update(value)
                out[key] = props
            elif key == "required" and isinstance(value, list):
                out[key] = sorted(set(out.get(key) or []) | set(value))
            else:
                out[key] = value
        return out

    def response_schema(self, op: dict[str, Any]) -> tuple[dict[str, Any], str]:
        responses = op.get("responses") or {}
        codes = sorted(
            (str(code), response)
            for code, response in responses.items()
            if str(code).startswith("2")
        )
        for code, raw in codes:
            response = self.expanded(raw)
            content = response.get("content") or {}
            for media in ("application/json", "application/vnd.api+json"):
                entry = content.get(media)
                if isinstance(entry, dict) and "schema" in entry:
                    return self.expanded(entry["schema"]), code
        return {}, ""

    def data_shape(self, op: dict[str, Any]) -> tuple[str, dict[str, Any], str]:
        schema, code = self.response_schema(op)
        props = schema.get("properties") or {}
        if "data" not in props:
            return "no-data", {}, code
        data = self.expanded(props["data"])
        if data.get("type") == "array" or "items" in data:
            return "array", self.expanded(data.get("items") or {}), code
        return "object", data, code

    def shape_fingerprint(self, data: dict[str, Any]) -> str:
        data = self.expanded(data)
        attrs = self.expanded((data.get("properties") or {}).get("attributes") or {})

        def structural(node: Any, depth: int = 0) -> Any:
            if depth > 10:
                return "<depth>"
            node = self.expanded(node) if isinstance(node, dict) else node
            if isinstance(node, list):
                return [structural(v, depth + 1) for v in node]
            if not isinstance(node, dict):
                return node
            keep: dict[str, Any] = {}
            for key in ("type", "format", "enum", "nullable", "oneOf", "anyOf", "items"):
                if key in node:
                    keep[key] = structural(node[key], depth + 1)
            props = node.get("properties") or {}
            if props:
                keep["properties"] = {
                    name: structural(raw, depth + 1) for name, raw in sorted(props.items())
                }
            if "additionalProperties" in node:
                keep["additionalProperties"] = structural(
                    node["additionalProperties"], depth + 1
                )
            return keep

        payload = json.dumps(structural(attrs), sort_keys=True, separators=(",", ":"))
        return hashlib.sha256(payload.encode()).hexdigest()


@dataclasses.dataclass
class Operation:
    operation_id: str
    path: str
    service: str
    kind: str
    data: dict[str, Any]
    annotated: bool
    annotated_cardinality: str = ""
    excluded_reason: str = ""
    scope_reason: str = ""
    spec_error: str = ""
    sdk_bound: Optional[bool] = None


@dataclasses.dataclass
class Resource:
    key: str
    display: str
    service: str
    by_id: Optional[Operation] = None
    singleton: Optional[Operation] = None
    collection: Optional[Operation] = None
    alternate_singular: list[Operation] = dataclasses.field(default_factory=list)

    @property
    def singular(self) -> Optional[Operation]:
        return self.by_id or self.singleton

    @property
    def endpoint_set(self) -> str:
        has_s = self.singular is not None
        has_p = self.collection is not None
        if has_s and has_p:
            return "S∩P"
        if has_s:
            return "S\\P"
        if has_p:
            return "P\\S"
        return "excluded"


@dataclasses.dataclass
class Candidate:
    candidate_id: str
    artifact_name: str
    resource: Resource
    cardinality: str
    read: Operation
    search: Optional[Operation]
    both_mode: str
    benefit: int
    annotated: bool
    sdk_bound: Optional[bool]
    pre_outcome: str = ""
    pre_diagnostic: str = ""

    @property
    def operations(self) -> list[Operation]:
        return [self.read] + ([self.search] if self.search else [])


@dataclasses.dataclass
class Result:
    candidate: Candidate
    outcome: str
    failure_class: str = ""
    gap_bucket: str = ""
    gap_stage: str = ""
    gap_confidence: str = ""
    diagnostic: str = ""
    warnings: list[str] = dataclasses.field(default_factory=list)


class SDKInventory:
    def __init__(self, repo: Path):
        self.version = ""
        self.operations: Optional[set[str]] = None
        go_mod = (repo / "go.mod").read_text(encoding="utf-8")
        match = re.search(
            r"github\.com/DataDog/datadog-api-client-go/v2\s+(\S+)", go_mod
        )
        if not match:
            return
        self.version = match.group(1)
        roots = []
        if os.environ.get("GOMODCACHE"):
            roots.append(Path(os.environ["GOMODCACHE"]))
        if os.environ.get("GOPATH"):
            roots.extend(Path(p) / "pkg/mod" for p in os.environ["GOPATH"].split(os.pathsep))
        roots.append(Path.home() / "go/pkg/mod")
        suffix = Path(
            "github.com/!data!dog/datadog-api-client-go"
        ) / f"v2@{self.version}" / "api/datadogV2"
        sdk_dir = next((root / suffix for root in roots if (root / suffix).is_dir()), None)
        if sdk_dir is None:
            return
        methods: set[str] = set()
        method_re = re.compile(r"^func \(a \*\w+Api\) ([A-Z][A-Za-z0-9]+)\(", re.M)
        for path in sdk_dir.glob("api_*.go"):
            methods.update(method_re.findall(path.read_text(encoding="utf-8", errors="ignore")))
        self.operations = methods

    def contains(self, operation_id: str) -> Optional[bool]:
        return None if self.operations is None else operation_id in self.operations


def canonical_resource_path(path: str, kind: str) -> str:
    clean = path.rstrip("/")
    if kind == "by-id":
        clean = re.sub(r"/\{[^/{}]+\}$", "", clean)
    return clean


def resource_display(path: str) -> str:
    parts = [
        snake(part)
        for part in path.strip("/").split("/")
        if part and not part.startswith("{") and part not in ("api", "v2", "unstable")
    ]
    return parts[-1] if parts else "unknown"


def is_by_id_path(path: str) -> bool:
    return bool(re.search(r"/\{[^/{}]+\}/?$", path))


def choose_artifact_name(operation_id: str, cardinality: str, used: set[str]) -> str:
    stem = snake(operation_id)
    suffix = "_plural" if cardinality == "plural" else "_singular"
    raw = GEN_PREFIX + stem + suffix
    if len(raw) > 64:
        digest = hashlib.sha1(raw.encode()).hexdigest()[:8]
        raw = raw[: 64 - len(digest) - 1].rstrip("_") + "_" + digest
    name = raw
    counter = 2
    while name in used:
        tail = f"_{counter}"
        name = raw[: 64 - len(tail)].rstrip("_") + tail
        counter += 1
    used.add(name)
    return name


def inventory(
    spec: dict[str, Any], sdk: SDKInventory
) -> tuple[list[Resource], list[Operation], dict[str, int]]:
    nav = SpecNavigator(spec)
    resources: dict[str, Resource] = {}
    excluded: list[Operation] = []
    raw_counts = {
        "get_operations": 0,
        "by_id": 0,
        "singleton": 0,
        "collection": 0,
        "no_data": 0,
    }
    for path, item in sorted((spec.get("paths") or {}).items()):
        if not isinstance(item, dict) or not isinstance(item.get(GET), dict):
            continue
        raw_counts["get_operations"] += 1
        op_node = item[GET]
        op_id = op_node.get("operationId") or ""
        service = (op_node.get("tags") or ["unclassified"])[0]
        shape, data, response_code = nav.data_shape(op_node)
        spec_error = ""
        if not op_id:
            spec_error = "GET operation is missing operationId"
        elif not response_code:
            spec_error = "GET operation has no JSON 2xx response schema"
        if shape == "array":
            kind = "collection"
        elif shape == "object" and is_by_id_path(path):
            kind = "by-id"
        else:
            kind = "singleton"
        if shape == "no-data":
            raw_counts["no_data"] += 1
        else:
            raw_counts[kind.replace("-", "_")] += 1
        reason = ""
        scope_reason = ""
        if path.startswith("/api/unstable/") or op_node.get("x-unstable"):
            scope_reason = "unstable/private preview surface"
        if re.match(r"^/api/v2/dashboards(?:/|$)", path):
            reason = "dashboards permanent exclusion"
        elif shape == "no-data":
            reason = "no JSON:API data envelope"
        extension = op_node.get("x-datadog-tf-generator")
        annotated = (
            isinstance(extension, dict)
            and extension.get("artifact_kind") == "data_source"
            and not extension.get("skip")
        )
        annotated_cardinality = ""
        if annotated:
            annotated_cardinality = extension.get("cardinality") or "singular"
        operation = Operation(
            operation_id=op_id or f"<missing:{path}>",
            path=path,
            service=service,
            kind=kind,
            data=data,
            annotated=annotated,
            annotated_cardinality=annotated_cardinality,
            excluded_reason=reason,
            scope_reason=scope_reason,
            spec_error=spec_error,
            sdk_bound=sdk.contains(op_id) if op_id else False,
        )
        if reason or spec_error:
            excluded.append(operation)
            continue
        key = canonical_resource_path(path, kind)
        resource = resources.setdefault(
            key,
            Resource(
                key=key,
                display=resource_display(key),
                service=service,
            ),
        )
        slot = kind.replace("-", "_")
        current = getattr(resource, slot)
        if current is None:
            setattr(resource, slot, operation)
        elif kind in ("by-id", "singleton"):
            resource.alternate_singular.append(operation)
        else:
            # One GET exists per exact path; retaining this as an explicit spec error
            # is safer than silently changing a denominator.
            operation.spec_error = f"duplicate {kind} operation for resource key {key}"
            excluded.append(operation)
    return list(resources.values()), excluded, raw_counts


def benefit_for(repo: Path, resource: Resource, cardinality: str) -> int:
    stem = singularize(resource.display)
    score = 40
    files = list((repo / "datadog/fwprovider").glob("resource_datadog_*.go"))
    files += list((repo / "datadog").glob("resource_datadog_*.go"))
    if any(stem in singularize(path.stem.removeprefix("resource_datadog_")) for path in files):
        score += 30
    if stem in {singularize(item) for item in HIGH_VALUE}:
        score += 15
    if cardinality == "plural":
        score += 10
    return min(score, 100)


def candidates_for(
    repo: Path, resources: list[Resource], nav: SpecNavigator
) -> list[Candidate]:
    candidates: list[Candidate] = []
    used: set[str] = set()
    for resource in sorted(resources, key=lambda r: r.key):
        singular = resource.singular
        collection = resource.collection
        both_mode = ""
        if singular and collection:
            both_mode = (
                "full"
                if nav.shape_fingerprint(singular.data)
                == nav.shape_fingerprint(collection.data)
                else "degraded"
            )
        if singular:
            artifact = choose_artifact_name(singular.operation_id, "singular", used)
            ops = [singular] + ([collection] if collection else [])
            sdk_bound = (
                None
                if any(op.sdk_bound is None for op in ops)
                else all(bool(op.sdk_bound) for op in ops)
            )
            candidates.append(
                Candidate(
                    candidate_id=f"{resource.key}#singular",
                    artifact_name=artifact,
                    resource=resource,
                    cardinality="singular",
                    read=singular,
                    search=collection,
                    both_mode=both_mode,
                    benefit=benefit_for(repo, resource, "singular"),
                    annotated=any(
                        op.annotated and op.annotated_cardinality == "singular"
                        for op in ops
                    ),
                    sdk_bound=sdk_bound,
                )
            )
        if collection:
            artifact = choose_artifact_name(collection.operation_id, "plural", used)
            candidates.append(
                Candidate(
                    candidate_id=f"{resource.key}#plural",
                    artifact_name=artifact,
                    resource=resource,
                    cardinality="plural",
                    read=collection,
                    search=None,
                    both_mode=both_mode,
                    benefit=benefit_for(repo, resource, "plural"),
                    annotated=(
                        collection.annotated
                        and collection.annotated_cardinality == "plural"
                    ),
                    sdk_bound=collection.sdk_bound,
                )
            )
    return candidates


def classify_gap(diagnostic: str, stage: str) -> tuple[str, str, str]:
    text = diagnostic.lower()
    patterns = (
        (r"map not yet supported|nesting under attributes.*shape=map", "maps-under-attributes"),
        (r"nesting under attributes", "nesting-under-attributes"),
        (r"missing an attributes object", "flat-data"),
        (r"single-member json:api envelope|no json:api data", "no-data-envelope"),
        (r"no result-array|missing results array block", "singleton-wrapper"),
        (r"id_strategy.+not yet supported", "id-strategy"),
        (r"uuid\.uuid|cannot use .+uuid.+ as string", "uuid-id"),
        (
            r"apiinstances\.get[a-z0-9]+apiv2 undefined",
            "api-accessor-resolution",
        ),
        (
            r"not enough arguments|too many arguments|cannot use optionalparams|"
            r"cannot use .+ as .+ in argument|cannot use .+ in assignment|"
            r"undefined: datadogv2\.[a-z0-9]+optionalparameters",
            "sdk-arg-binding",
        ),
        (
            r"redeclared|other declaration|duplicate tfsdk|declares tfsdk:.+twice",
            "id-collision",
        ),
        (
            r"gofmt of generated data source.+expected operand, found 'type'",
            "reserved-identifier",
        ),
        (
            r"circular \$ref|schema kind \"ref_cycle\"|"
            r"(?:array element|map value) kind \"ref_cycle\"",
            "recursive-schema",
        ),
        (
            r"\$ref expansion exceeded --max-depth",
            "schema-depth-limit",
        ),
        (
            r"schema kind \"unsupported\"|"
            r"(?:array element|map value) kind \"unsupported\"",
            "unsupported-schema-kind",
        ),
        (r"no read or search sdk call resolved", "sdk-call-resolution"),
        (r"datadogunstable|no required module provides package", "sdk-not-released"),
    )
    for pattern, bucket in patterns:
        if re.search(pattern, text, re.S):
            confidence = "high" if stage == "emit" else "medium"
            return bucket, stage, confidence
    return "unclassified", stage, "low"


class Slicer:
    def __init__(self, repo: Path, spec_path: Path, spec: dict[str, Any], temp: Path):
        scripts = repo / ".generator-v2/internal/testdata/mini-oas/scripts"
        sys.path.insert(0, str(scripts))
        from _build_mini import build_slice  # type: ignore

        self.build_slice = build_slice
        self.spec = spec
        self.spec_path = spec_path
        self.temp = temp
        self.index: dict[str, tuple[str, str]] = {}
        for path, item in (spec.get("paths") or {}).items():
            if not isinstance(item, dict):
                continue
            for method, node in item.items():
                if isinstance(node, dict) and node.get("operationId"):
                    self.index[node["operationId"]] = (path, method)
        self.security_schemes = copy.deepcopy(
            (spec.get("components") or {}).get("securitySchemes") or {}
        )

    def create(self, candidate: Candidate) -> Path:
        ops = [candidate.read.operation_id]
        group = {"read": candidate.read.operation_id}
        if candidate.search:
            ops.append(candidate.search.operation_id)
            group["search"] = candidate.search.operation_id
        ext: dict[str, Any] = {
            "artifact_kind": "data_source",
            "artifact_name": candidate.artifact_name,
            "tf_description": (
                "Coverage-sweep probe. This generated artifact is always discarded."
            ),
            "group": group,
        }
        if candidate.cardinality == "plural":
            ext["cardinality"] = "plural"
        sentinel = object()
        saved_annotations: list[tuple[dict[str, Any], Any]] = []
        # The source bundle may already carry shipping annotations. A candidate
        # slice can include more than its anchor operation (singular read +
        # collection search), so suppress every original annotation first. If we
        # leave one in place, tfgen emits a real artifact alongside the disposable
        # coverage probe and invalidates both attribution and cleanup.
        for operation_id in ops:
            op_path, op_method = self.index[operation_id]
            op_node = self.spec["paths"][op_path][op_method]
            saved_annotations.append(
                (op_node, op_node.get("x-datadog-tf-generator", sentinel))
            )
            op_node.pop("x-datadog-tf-generator", None)
        path, method = self.index[candidate.read.operation_id]
        node = self.spec["paths"][path][method]
        node["x-datadog-tf-generator"] = ext
        out = self.temp / f"{candidate.artifact_name}.yaml"
        try:
            self.build_slice(self.spec, self.index, ops, str(out))
        finally:
            for op_node, old in saved_annotations:
                if old is sentinel:
                    op_node.pop("x-datadog-tf-generator", None)
                else:
                    op_node["x-datadog-tf-generator"] = old
            if self.security_schemes:
                self.spec.setdefault("components", {})["securitySchemes"] = copy.deepcopy(
                    self.security_schemes
                )
        return out


class TreeGuard:
    def __init__(self, repo: Path):
        self.repo = repo
        self.start_status = self._status()
        target = run_command(
            ["git", "status", "--porcelain", "--", *TARGET_PATHS], repo
        ).stdout.strip()
        if target:
            raise RuntimeError(
                "coverage sweep refuses to generate onto dirty target paths:\n" + target
            )

    def _status(self) -> str:
        return run_command(["git", "status", "--porcelain=v1", "-z"], self.repo).stdout

    def cleanup(self) -> None:
        registry = self.repo / "datadog/fwprovider/datasources_generated.go"
        run_command(
            ["git", "restore", "--worktree", "--", str(registry.relative_to(self.repo))],
            self.repo,
        )
        for path in (self.repo / "datadog/fwprovider").glob(
            f"data_source_datadog_{GEN_PREFIX}*.go"
        ):
            path.unlink()
        examples = self.repo / "examples/data-sources"
        for path in examples.glob(f"datadog_{GEN_PREFIX}*"):
            if path.is_dir():
                shutil.rmtree(path)
            else:
                path.unlink()

    def assert_restored(self) -> None:
        self.cleanup()
        end = self._status()
        if end != self.start_status:
            raise RuntimeError(
                "coverage sweep did not restore the starting git status\n"
                f"before={self.start_status!r}\nafter={end!r}"
            )


def parse_emit_report(candidate: Candidate, report_path: Path) -> Result:
    if not report_path.exists():
        return Result(
            candidate,
            "emit-fail",
            "generator-gap",
            "unclassified",
            "emit",
            "low",
            "tfgen did not write a report",
        )
    report = json.loads(report_path.read_text(encoding="utf-8"))
    artifacts = [
        item
        for item in report.get("artifacts") or []
        if item.get("name") == candidate.artifact_name
        and item.get("kind") == "data_source"
    ]
    if not artifacts:
        return Result(
            candidate,
            "emit-fail",
            "generator-gap",
            "unclassified",
            "emit",
            "low",
            "report has no Go data-source artifact entry",
        )
    artifact = next(
        (
            item
            for item in artifacts
            if str(item.get("path") or "").endswith(".go")
            or item.get("status") == "failed"
        ),
        artifacts[0],
    )
    diagnostics = artifact.get("diagnostics") or []
    errors = [d.get("message", "") for d in diagnostics if d.get("severity") == "error"]
    warnings = [
        d.get("message", "")
        for d in diagnostics
        if d.get("severity") in ("warning", "info")
    ]
    if artifact.get("status") == "failed" or errors:
        diagnostic = "\n".join(errors) or f"artifact status={artifact.get('status')}"
        bucket, gap_stage, confidence = classify_gap(diagnostic, "emit")
        return Result(
            candidate,
            "emit-fail",
            "generator-gap",
            bucket,
            gap_stage,
            confidence,
            diagnostic,
            warnings,
        )
    if artifact.get("status") not in ("created", "updated", "unchanged"):
        diagnostic = f"unexpected artifact status={artifact.get('status')}"
        return Result(
            candidate,
            "emit-fail",
            "generator-gap",
            "unclassified",
            "emit",
            "low",
            diagnostic,
            warnings,
        )
    return Result(candidate, "emit-pass", warnings=warnings)


SOURCE_ERROR_RE = re.compile(
    rf"(?P<file>(?:\./)?datadog/fwprovider/data_source_datadog_"
    rf"(?P<name>{GEN_PREFIX}[a-z0-9_]+)\.go):(?P<line>\d+):(?P<col>\d+):\s*(?P<msg>.+)"
)
REGISTRY_ERROR_RE = re.compile(
    r"(?P<file>(?:\./)?datadog/fwprovider/datasources_generated\.go):"
    r"(?P<line>\d+):(?P<col>\d+):\s*(?P<msg>.+)"
)


def attribute_build_errors(
    output: str, candidates: list[Candidate], repo: Path
) -> dict[str, str]:
    by_name = {candidate.artifact_name: candidate for candidate in candidates}
    errors: dict[str, list[str]] = {}
    for match in SOURCE_ERROR_RE.finditer(output):
        if match.group("name") in by_name:
            errors.setdefault(match.group("name"), []).append(match.group("msg").strip())
    registry = repo / "datadog/fwprovider/datasources_generated.go"
    lines = registry.read_text(encoding="utf-8").splitlines() if registry.exists() else []
    constructor_map = {
        "New" + "".join(part.title() for part in ("datadog_" + c.artifact_name).split("_"))
        + "DataSource": c.artifact_name
        for c in candidates
    }
    for match in REGISTRY_ERROR_RE.finditer(output):
        line_number = int(match.group("line"))
        source_line = lines[line_number - 1] if 0 < line_number <= len(lines) else ""
        matched_name = next(
            (name for constructor, name in constructor_map.items() if constructor in source_line),
            None,
        )
        if matched_name:
            errors.setdefault(matched_name, []).append(match.group("msg").strip())
    return {name: "\n".join(messages) for name, messages in errors.items()}


def generate_candidate(
    repo: Path,
    tfgen: Path,
    slicer: Slicer,
    candidate: Candidate,
    reports: Path,
) -> Result:
    spec_slice = slicer.create(candidate)
    report_path = reports / f"{candidate.artifact_name}.json"
    proc = run_command(
        [str(tfgen), "generate", "--spec", str(spec_slice), "--report", str(report_path)],
        repo,
        timeout=180,
    )
    result = parse_emit_report(candidate, report_path)
    if (
        result.outcome == "emit-fail"
        and result.diagnostic == "tfgen did not write a report"
        and proc.stdout.strip()
    ):
        diagnostic = proc.stdout.strip()
        bucket, stage, confidence = classify_gap(diagnostic, "emit")
        result = Result(
            candidate,
            "emit-fail",
            "generator-gap",
            bucket,
            stage,
            confidence,
            diagnostic,
        )
    if result.outcome == "emit-pass" and proc.returncode:
        result = Result(
            candidate,
            "emit-fail",
            "generator-gap",
            "unclassified",
            "emit",
            "low",
            proc.stdout.strip() or f"tfgen exited {proc.returncode}",
            result.warnings,
        )
    return result


def sweep(
    repo: Path,
    spec_path: Path,
    spec: dict[str, Any],
    candidates: list[Candidate],
    output_dir: Path,
    skip_baseline_build: bool,
    build_timeout: int,
) -> list[Result]:
    tfgen = repo / "bin/tfgen"
    # A coverage result is only meaningful for the checked-out generator.  An
    # existing binary may have been built by another branch, so always replace
    # it before probing candidates.
    eprint("building current tfgen with make tfgen-build")
    run_command(["make", "tfgen-build"], repo, timeout=build_timeout, check=True)
    guard = TreeGuard(repo)
    results: dict[str, Result] = {}
    reports = output_dir / "tfgen-reports"
    reports.mkdir(parents=True, exist_ok=True)
    try:
        if not skip_baseline_build:
            eprint("checking baseline with make build")
            baseline = run_command(["make", "build"], repo, timeout=build_timeout)
            (output_dir / "baseline-build.log").write_text(
                baseline.stdout, encoding="utf-8"
            )
            if baseline.returncode:
                raise RuntimeError(
                    "baseline make build failed; refusing to attribute existing failures "
                    "to coverage candidates\n" + baseline.stdout
                )
        with tempfile.TemporaryDirectory(prefix="tfgen-coverage-") as temp_name:
            slicer = Slicer(repo, spec_path, spec, Path(temp_name))
            emit_clean: list[Candidate] = []
            total = len(candidates)
            for index, candidate in enumerate(candidates, 1):
                if candidate.pre_outcome:
                    results[candidate.candidate_id] = Result(
                        candidate,
                        candidate.pre_outcome,
                        (
                            "sdk-not-released"
                            if candidate.pre_outcome == "sdk-not-released"
                            else candidate.pre_outcome
                        ),
                        candidate.pre_outcome,
                        "inventory",
                        "high",
                        candidate.pre_diagnostic,
                    )
                    continue
                eprint(
                    f"emit {index}/{total} {candidate.cardinality} "
                    f"{candidate.read.operation_id}"
                )
                result = generate_candidate(repo, tfgen, slicer, candidate, reports)
                results[candidate.candidate_id] = result
                guard.cleanup()
                if result.outcome == "emit-pass":
                    emit_clean.append(candidate)

            pending_batches = [emit_clean] if emit_clean else []
            round_number = 0
            max_build_rounds = max(30, len(emit_clean) * 2)
            while pending_batches:
                pending = pending_batches.pop(0)
                round_number += 1
                if round_number > max_build_rounds:
                    raise RuntimeError(
                        f"build attribution exceeded {max_build_rounds} rounds"
                    )
                guard.cleanup()
                eprint(f"build round {round_number}: {len(pending)} candidates")
                for candidate in pending:
                    regenerated = generate_candidate(
                        repo, tfgen, slicer, candidate, reports
                    )
                    if regenerated.outcome != "emit-pass":
                        results[candidate.candidate_id] = regenerated
                pending = [
                    candidate
                    for candidate in pending
                    if results[candidate.candidate_id].outcome == "emit-pass"
                ]
                if not pending:
                    continue
                build = run_command(
                    ["make", "build"], repo, timeout=build_timeout
                )
                (output_dir / f"build-round-{round_number}.log").write_text(
                    build.stdout, encoding="utf-8"
                )
                if build.returncode == 0:
                    for candidate in pending:
                        prior = results[candidate.candidate_id]
                        results[candidate.candidate_id] = Result(
                            candidate, "pass", warnings=prior.warnings
                        )
                    continue
                attributed = attribute_build_errors(build.stdout, pending, repo)
                if attributed:
                    survivors: list[Candidate] = []
                    collision_retries: list[Candidate] = []
                    for candidate in pending:
                        diagnostic = attributed.get(candidate.artifact_name)
                        if diagnostic is None:
                            survivors.append(candidate)
                            continue
                        bucket, stage, confidence = classify_gap(diagnostic, "build")
                        # Package-level model names are not artifact-scoped. A batch
                        # can therefore create a redeclaration that neither candidate
                        # has when compiled alone. Recursively separate collision
                        # candidates; only a singleton collision is a real failure.
                        if bucket == "id-collision" and len(pending) > 1:
                            collision_retries.append(candidate)
                            results[candidate.candidate_id].warnings.append(
                                "retrying after batch-only identifier collision"
                            )
                            continue
                        failure_class = (
                            "sdk-not-released"
                            if bucket == "sdk-not-released"
                            else "generator-gap"
                        )
                        outcome = (
                            "sdk-not-released"
                            if bucket == "sdk-not-released"
                            else "build-fail"
                        )
                        results[candidate.candidate_id] = Result(
                            candidate,
                            outcome,
                            failure_class,
                            bucket,
                            stage,
                            confidence,
                            diagnostic,
                            results[candidate.candidate_id].warnings,
                        )
                    if survivors:
                        pending_batches.insert(0, survivors)
                    if collision_retries:
                        midpoint = max(1, len(collision_retries) // 2)
                        first = collision_retries[:midpoint]
                        second = collision_retries[midpoint:]
                        if second:
                            pending_batches.insert(0, second)
                        pending_batches.insert(0, first)
                    continue
                if len(pending) == 1:
                    candidate = pending[0]
                    results[candidate.candidate_id] = Result(
                        candidate,
                        "build-fail",
                        "generator-gap",
                        "unclassified",
                        "build",
                        "low",
                        build.stdout[-12000:],
                        results[candidate.candidate_id].warnings,
                    )
                else:
                    # An unattributed package-level error is isolated with a smaller
                    # batch. Candidates in the other half are revisited afterward.
                    midpoint = max(1, len(pending) // 2)
                    first, second = pending[:midpoint], pending[midpoint:]
                    for candidate in second:
                        results[candidate.candidate_id].warnings.append(
                            "deferred during unattributed build-error bisection"
                        )
                    pending_batches.insert(0, second)
                    pending_batches.insert(0, first)
    finally:
        guard.assert_restored()
    return [results[c.candidate_id] for c in candidates]


def derived_both_results(
    resources: list[Resource], candidates: list[Candidate], results: list[Result]
) -> list[Result]:
    candidate_by_key = {
        (candidate.resource.key, candidate.cardinality): candidate
        for candidate in candidates
    }
    result_by_id = {result.candidate.candidate_id: result for result in results}
    derived: list[Result] = []
    for resource in resources:
        if resource.endpoint_set != "S∩P":
            continue
        singular = candidate_by_key.get((resource.key, "singular"))
        plural = candidate_by_key.get((resource.key, "plural"))
        if singular is None or plural is None:
            continue
        s_result = result_by_id[singular.candidate_id]
        p_result = result_by_id[plural.candidate_id]
        proxy = dataclasses.replace(
            singular,
            candidate_id=f"{resource.key}#both",
            artifact_name="",
            cardinality="both",
            read=singular.read,
            search=plural.read,
            benefit=round((singular.benefit + plural.benefit) / 2),
            annotated=singular.annotated and plural.annotated,
            sdk_bound=(
                None
                if singular.sdk_bound is None or plural.sdk_bound is None
                else singular.sdk_bound and plural.sdk_bound
            ),
        )
        if s_result.outcome == "pass" and p_result.outcome == "pass":
            derived.append(Result(proxy, "pass"))
            continue
        order = {
            "sdk-not-released": 0,
            "out-of-scope": 1,
            "spec-error": 2,
            "emit-fail": 3,
            "build-fail": 4,
            "not-run": 5,
        }
        failures = [r for r in (s_result, p_result) if r.outcome != "pass"]
        chosen = sorted(failures, key=lambda r: order.get(r.outcome, 99))[0]
        diagnostic = (
            f"singular={s_result.outcome}; plural={p_result.outcome}; "
            f"{chosen.diagnostic}"
        ).strip()
        derived.append(
            Result(
                proxy,
                chosen.outcome,
                chosen.failure_class,
                chosen.gap_bucket,
                chosen.gap_stage,
                chosen.gap_confidence,
                diagnostic,
            )
        )
    return derived


def result_row(
    result: Result, run_date: str, commit: str, sdk_version: str
) -> dict[str, Any]:
    candidate = result.candidate
    operation_id = candidate.read.operation_id
    if candidate.search:
        operation_id += "+" + candidate.search.operation_id
    return {
        "run_date": run_date,
        "api_spec_commit": commit,
        "sdk_version": sdk_version,
        "service": candidate.resource.service,
        "resource": candidate.resource.display,
        "resource_path": candidate.resource.key,
        "operation_id": operation_id,
        "cardinality": candidate.cardinality,
        "endpoint_set": candidate.resource.endpoint_set,
        "both_mode": candidate.both_mode,
        "annotated": str(candidate.annotated).lower(),
        "sdk_bound": (
            "unknown" if candidate.sdk_bound is None else str(candidate.sdk_bound).lower()
        ),
        "outcome": result.outcome,
        "failure_class": result.failure_class,
        "gap_bucket": result.gap_bucket,
        "gap_stage": result.gap_stage,
        "gap_confidence": result.gap_confidence,
        "gap_diagnostic": result.diagnostic.replace("\n", " | "),
        "benefit": candidate.benefit,
        "value_weighted": candidate.benefit,
    }


def write_report(
    output_dir: Path,
    resources: list[Resource],
    excluded: list[Operation],
    raw_counts: dict[str, int],
    results: list[Result],
    both: list[Result],
    metadata: dict[str, Any],
) -> tuple[Path, Path, Path]:
    all_results = results + both
    run_date = metadata["run_date"]
    rows = [
        result_row(r, run_date, metadata["api_spec_commit"], metadata["sdk_version"])
        for r in all_results
    ]
    csv_path = output_dir / "coverage.csv"
    with csv_path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=CSV_FIELDS)
        writer.writeheader()
        writer.writerows(rows)

    set_counts = {"S∩P": 0, "S\\P": 0, "P\\S": 0}
    for resource in resources:
        set_counts[resource.endpoint_set] += 1

    def view(cardinality: str, subset: Optional[str] = None) -> dict[str, Any]:
        selected = [
            r
            for r in all_results
            if r.candidate.cardinality == cardinality
            and (subset is None or r.candidate.both_mode == subset)
        ]
        in_scope = [
            r
            for r in selected
            if r.outcome not in ("out-of-scope", "spec-error", "excluded")
        ]
        sdk_bound = [
            r
            for r in in_scope
            if r.candidate.sdk_bound is not False
        ]
        passed = [r for r in sdk_bound if r.outcome == "pass"]
        weighted_den = sum(r.candidate.benefit for r in sdk_bound)
        weighted_num = sum(r.candidate.benefit for r in passed)
        not_run = sum(r.outcome == "not-run" for r in sdk_bound)
        return {
            "inventory": len(selected),
            "full_surface": len(in_scope),
            "full_surface_coverage": (
                "n/a (not run)"
                if not_run
                else percent(len(passed), len(in_scope))
            ),
            "sdk_bound_denominator": len(sdk_bound),
            "pass": len(passed),
            "coverage": "n/a (not run)" if not_run else percent(len(passed), len(sdk_bound)),
            "weighted_coverage": (
                "n/a (not run)" if not_run else percent(weighted_num, weighted_den)
            ),
            "sdk_not_released": sum(r.outcome == "sdk-not-released" for r in selected),
            "out_of_scope": sum(r.outcome == "out-of-scope" for r in selected),
            "spec_error": sum(r.outcome == "spec-error" for r in selected),
            "emit_fail": sum(r.outcome == "emit-fail" for r in selected),
            "build_fail": sum(r.outcome == "build-fail" for r in selected),
            "not_run": not_run,
        }

    views = {
        "singular": view("singular"),
        "plural": view("plural"),
        "both": view("both"),
        "both_full": view("both", "full"),
        "both_degraded": view("both", "degraded"),
    }
    intent = [r for r in all_results if r.candidate.annotated]
    intent_denominator = sum(
        r.candidate.sdk_bound is not False
        and r.outcome not in ("out-of-scope", "spec-error", "excluded")
        for r in intent
    )
    intent_not_run = sum(r.outcome == "not-run" for r in intent)
    gaps: dict[str, dict[str, Any]] = {}
    for result in all_results:
        if result.failure_class != "generator-gap":
            continue
        bucket = result.gap_bucket or "unclassified"
        entry = gaps.setdefault(
            bucket, {"count": 0, "singular": 0, "plural": 0, "both": 0}
        )
        entry["count"] += 1
        entry[result.candidate.cardinality] += 1
    json_payload = {
        "metadata": metadata,
        "raw_operation_counts": raw_counts,
        "set_counts": set_counts,
        "coverage_views": views,
        "intent": {
            "rows": len(intent),
            "pass": sum(r.outcome == "pass" for r in intent),
            "coverage": (
                "n/a (not run)"
                if intent_not_run
                else percent(
                    sum(r.outcome == "pass" for r in intent),
                    intent_denominator,
                )
            ),
        },
        "gap_histogram": dict(
            sorted(gaps.items(), key=lambda item: (-item[1]["count"], item[0]))
        ),
        "exclusions": [dataclasses.asdict(op) for op in excluded],
        "rows": rows,
    }
    json_path = output_dir / "coverage.json"
    json_path.write_text(json.dumps(json_payload, indent=2), encoding="utf-8")

    lines = [
        "# tfgen v2 data-source coverage sweep",
        "",
        f"- Run: `{run_date}`",
        f"- API spec: `{metadata['spec_path']}` @ `{metadata['api_spec_commit']}`",
        f"- Provider: `{metadata['provider_commit']}`",
        f"- SDK: `{metadata['sdk_version'] or 'unknown'}`",
        f"- Mode: `{'enumerate-only' if metadata['enumerate_only'] else 'build-verified'}`",
        "",
        "## OAS-derived endpoint sets",
        "",
        "| View | Count |",
        "|---|---:|",
        f"| GET operations inspected | {raw_counts['get_operations']} |",
        f"| by-id object reads | {raw_counts['by_id']} |",
        f"| singleton object reads (counted in S, never P) | {raw_counts['singleton']} |",
        f"| true array collections (P) | {raw_counts['collection']} |",
        f"| responses without a data envelope | {raw_counts['no_data']} |",
        f"| S∩P | {set_counts['S∩P']} |",
        "| S\\P | {} |".format(set_counts["S\\P"]),
        "| P\\S | {} |".format(set_counts["P\\S"]),
        f"| permanent/spec exclusions | {len(excluded)} |",
        "",
        "## Build-verified coverage",
        "",
        "The headline denominator is SDK-bound candidates. Full-surface size and "
        "`sdk-not-released` remain visible separately.",
        "",
        "| Cardinality | Pass | SDK-bound denominator | Headline coverage | Value-weighted | In-scope spec surface | Spec-surface coverage | Total OAS inventory | SDK not released | Out of scope | Emit fail | Build fail |",
        "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|",
    ]
    for name in ("singular", "plural", "both", "both_full", "both_degraded"):
        v = views[name]
        lines.append(
            f"| {name.replace('_', ' ')} | {v['pass']} | "
            f"{v['sdk_bound_denominator']} | {v['coverage']} | "
            f"{v['weighted_coverage']} | {v['full_surface']} | "
            f"{v['full_surface_coverage']} | {v['inventory']} | "
            f"{v['sdk_not_released']} | {v['out_of_scope']} | "
            f"{v['emit_fail']} | {v['build_fail']} |"
        )
    lines += [
        "",
        "## Intent vs capability",
        "",
        f"- Total OAS inventory rows (including out-of-scope): {len(all_results)}.",
        f"- Annotated intent rows: {len(intent)}; build-verified pass "
        f"{sum(r.outcome == 'pass' for r in intent)}; coverage "
        f"{json_payload['intent']['coverage']}.",
        "",
        "## Generator-gap histogram",
        "",
        "| Gap | Count | Singular | Plural | Both | Singular-point lift | Plural-point lift |",
        "|---|---:|---:|---:|---:|---:|---:|",
    ]
    singular_den = views["singular"]["sdk_bound_denominator"]
    plural_den = views["plural"]["sdk_bound_denominator"]
    for bucket, entry in json_payload["gap_histogram"].items():
        lines.append(
            f"| {bucket} | {entry['count']} | {entry['singular']} | "
            f"{entry['plural']} | {entry['both']} | "
            f"{percent(entry['singular'], singular_den)} | "
            f"{percent(entry['plural'], plural_den)} |"
        )
    if not gaps:
        lines.append("| _none / not run_ | 0 | 0 | 0 | 0 | n/a | n/a |")
    unclassified = gaps.get("unclassified", {}).get("count", 0)
    lines += [
        "",
        f"**Unclassified generator gaps: {unclassified}.**",
        "",
        "## Permanent exclusions",
        "",
        "| Service | Operation | Path | Reason |",
        "|---|---|---|---|",
    ]
    for op in excluded:
        reason = op.spec_error or op.excluded_reason
        lines.append(f"| {op.service} | `{op.operation_id}` | `{op.path}` | {reason} |")
    if not excluded:
        lines.append("| _none_ | | | |")
    lines += [
        "",
        "## Artifacts",
        "",
        f"- Per-candidate CSV: `{csv_path}`",
        f"- Machine-readable report: `{json_path}`",
        f"- tfgen and build logs: `{output_dir}`",
        "",
        "“Pass” means generated and compiled; it does not claim live API or cassette validation.",
        "",
    ]
    report_path = output_dir / "coverage.md"
    report_path.write_text("\n".join(lines), encoding="utf-8")
    return csv_path, report_path, json_path


def copy_latest(output_dir: Path, results_root: Path) -> None:
    for source, target in (
        (output_dir / "coverage.csv", results_root / "latest.csv"),
        (output_dir / "coverage.md", results_root / "latest.md"),
        (output_dir / "coverage.json", results_root / "latest.json"),
    ):
        shutil.copy2(source, target)


def parse_args() -> argparse.Namespace:
    script_dir = Path(__file__).resolve().parent
    default_repo = script_dir.parents[2]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--provider-repo", type=Path, default=default_repo)
    parser.add_argument(
        "--api-spec-repo",
        type=Path,
        default=Path("/Users/jason.tenczar/projects/datadog-api-spec"),
    )
    parser.add_argument("--spec", type=Path)
    parser.add_argument("--fast", action="store_true")
    parser.add_argument("--enumerate-only", action="store_true")
    parser.add_argument("--limit", type=int, default=0)
    parser.add_argument("--candidate-regex", default="")
    parser.add_argument("--skip-baseline-build", action="store_true")
    parser.add_argument("--build-timeout", type=int, default=1800)
    parser.add_argument("--output-dir", type=Path)
    return parser.parse_args()


def git_value(repo: Path, *args: str) -> str:
    return run_command(["git", *args], repo, check=True).stdout.strip()


def main() -> int:
    args = parse_args()
    repo = args.provider_repo.resolve()
    api_repo = args.api_spec_repo.resolve()
    if args.spec:
        spec_path = args.spec.resolve()
    else:
        name = "full_spec.terraform_slim.yaml" if args.fast else "full_spec.terraform.yaml"
        spec_path = api_repo / "spec/v2" / name
    if not spec_path.is_file():
        raise RuntimeError(f"spec does not exist: {spec_path}")
    results_root = Path(__file__).resolve().parent / "results"
    run_id = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    output_dir = (
        args.output_dir.resolve()
        if args.output_dir
        else results_root / run_id
    )
    output_dir.mkdir(parents=True, exist_ok=False)
    eprint(f"loading {spec_path}")
    with spec_path.open(encoding="utf-8") as handle:
        spec = yaml.safe_load(handle)
    sdk = SDKInventory(repo)
    resources, excluded, raw_counts = inventory(spec, sdk)
    candidates = candidates_for(repo, resources, SpecNavigator(spec))
    for candidate in candidates:
        scope_reasons = sorted(
            {op.scope_reason for op in candidate.operations if op.scope_reason}
        )
        if scope_reasons:
            candidate.pre_outcome = "out-of-scope"
            candidate.pre_diagnostic = "; ".join(scope_reasons)
        elif candidate.sdk_bound is False:
            candidate.pre_outcome = "sdk-not-released"
            missing = [op.operation_id for op in candidate.operations if op.sdk_bound is False]
            candidate.pre_diagnostic = (
                f"operation(s) absent from released Go SDK {sdk.version}: "
                + ", ".join(missing)
            )
    if args.candidate_regex:
        pattern = re.compile(args.candidate_regex)
        candidates = [
            candidate
            for candidate in candidates
            if pattern.search(candidate.candidate_id)
            or pattern.search(candidate.read.operation_id)
        ]
        selected_keys = {candidate.resource.key for candidate in candidates}
        resources = [r for r in resources if r.key in selected_keys]
    if args.limit:
        candidates = candidates[: args.limit]
        selected_keys = {candidate.resource.key for candidate in candidates}
        resources = [r for r in resources if r.key in selected_keys]
    metadata = {
        "run_date": dt.date.today().isoformat(),
        "run_id": run_id,
        "spec_path": str(spec_path),
        "api_spec_commit": git_value(api_repo, "rev-parse", "HEAD"),
        "provider_commit": git_value(repo, "rev-parse", "HEAD"),
        "provider_branch": git_value(repo, "rev-parse", "--abbrev-ref", "HEAD"),
        "sdk_version": sdk.version,
        "sdk_inventory_available": sdk.operations is not None,
        "enumerate_only": args.enumerate_only,
        "fast": args.fast,
        "candidate_regex": args.candidate_regex,
        "limit": args.limit,
    }
    eprint(
        f"derived {len(resources)} resources and {len(candidates)} probe candidates; "
        f"excluded {len(excluded)} operations"
    )
    if args.enumerate_only:
        results = [
            Result(
                candidate,
                candidate.pre_outcome or "not-run",
                candidate.pre_outcome,
                candidate.pre_outcome,
                "inventory" if candidate.pre_outcome else "",
                "high" if candidate.pre_outcome else "",
                candidate.pre_diagnostic,
            )
            for candidate in candidates
        ]
    else:
        results = sweep(
            repo,
            spec_path,
            spec,
            candidates,
            output_dir,
            args.skip_baseline_build,
            args.build_timeout,
        )
    both = derived_both_results(resources, candidates, results)
    csv_path, report_path, json_path = write_report(
        output_dir,
        resources,
        excluded,
        raw_counts,
        results,
        both,
        metadata,
    )
    results_root.mkdir(parents=True, exist_ok=True)
    copy_latest(output_dir, results_root)
    print(report_path.read_text(encoding="utf-8"))
    print(f"\nCSV: {csv_path}")
    print(f"Report: {report_path}")
    print(f"JSON: {json_path}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        eprint("interrupted")
        raise SystemExit(130)
    except Exception as exc:
        eprint(f"coverage sweep failed: {exc}")
        raise SystemExit(1)
