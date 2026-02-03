#!/usr/bin/env python3
"""
FDL Skeleton Generator
Generates skeleton code from FDL (Feature Definition Language) YAML files.

Usage:
    python3 skeleton_generator.py --fdl <file> --output-dir <dir> [options]

Options:
    --fdl           FDL file path (required)
    --output-dir    Output directory (default: .)
    --backend       Backend type: python, go, node (default: python)
    --frontend      Frontend type: react, vue, none (default: none)
    --force         Overwrite existing files
    --dry-run       Show files that would be generated
"""

import argparse
import hashlib
import json
import os
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    import yaml
except ImportError:
    print(json.dumps({"generated_files": [], "errors": ["PyYAML not installed. Run: pip install pyyaml"]}))
    sys.exit(1)


@dataclass
class FDLField:
    name: str
    type: str
    constraints: str = ""


@dataclass
class FDLModel:
    name: str
    table: str
    fields: List[FDLField] = field(default_factory=list)


@dataclass
class FDLService:
    name: str
    desc: str = ""
    input: Dict[str, Any] = field(default_factory=dict)
    output: Optional[str] = None
    steps: List[str] = field(default_factory=list)


@dataclass
class FDLAPI:
    path: str
    method: str
    use: str = ""
    summary: str = ""
    request: Dict[str, Any] = field(default_factory=dict)
    response: Dict[str, Any] = field(default_factory=dict)


@dataclass
class FDLUI:
    component: str
    type: str = "Organism"
    props: Dict[str, Any] = field(default_factory=dict)
    state: List[str] = field(default_factory=list)
    init: List[str] = field(default_factory=list)


@dataclass
class FDLSpec:
    feature: str
    description: str = ""
    models: List[FDLModel] = field(default_factory=list)
    service: List[FDLService] = field(default_factory=list)
    api: List[FDLAPI] = field(default_factory=list)
    ui: List[FDLUI] = field(default_factory=list)


def parse_fdl(file_path: str) -> FDLSpec:
    """Parse FDL YAML file."""
    with open(file_path, 'r', encoding='utf-8') as f:
        data = yaml.safe_load(f)

    spec = FDLSpec(
        feature=data.get('feature', ''),
        description=data.get('description', '')
    )

    # Parse models
    for m in data.get('models', []):
        fields = []
        for f in m.get('fields', []):
            if isinstance(f, dict):
                fields.append(FDLField(
                    name=f.get('name', ''),
                    type=f.get('type', ''),
                    constraints=f.get('constraints', '')
                ))
        spec.models.append(FDLModel(
            name=m.get('name', ''),
            table=m.get('table', ''),
            fields=fields
        ))

    # Parse services
    for s in data.get('service', []):
        spec.service.append(FDLService(
            name=s.get('name', ''),
            desc=s.get('desc', ''),
            input=s.get('input', {}),
            output=s.get('output'),
            steps=s.get('steps', [])
        ))

    # Parse APIs
    for a in data.get('api', []):
        spec.api.append(FDLAPI(
            path=a.get('path', ''),
            method=a.get('method', ''),
            use=a.get('use', ''),
            summary=a.get('summary', ''),
            request=a.get('request', {}),
            response=a.get('response', {})
        ))

    # Parse UIs
    for u in data.get('ui', []):
        spec.ui.append(FDLUI(
            component=u.get('component', ''),
            type=u.get('type', 'Organism'),
            props=u.get('props', {}),
            state=u.get('state', []),
            init=u.get('init', [])
        ))

    return spec


def type_to_python(fdl_type: str) -> str:
    """Convert FDL type to Python/SQLAlchemy type."""
    type_map = {
        'uuid': 'String',
        'string': 'String',
        'text': 'Text',
        'int': 'Integer',
        'integer': 'Integer',
        'float': 'Float',
        'bool': 'Boolean',
        'boolean': 'Boolean',
        'datetime': 'DateTime',
        'date': 'Date',
        'json': 'JSON',
    }
    return type_map.get(fdl_type.lower(), 'String')


def generate_model_python(model: FDLModel, output_dir: str) -> str:
    """Generate Python SQLAlchemy model."""
    lines = [
        '"""',
        f'{model.name} Model',
        'Auto-generated from FDL. DO NOT modify structure.',
        '"""',
        'from sqlalchemy import Column, String, Integer, DateTime, Boolean, Text, ForeignKey',
        'from sqlalchemy.sql import func',
        'from app.db import Base',
        '',
        '',
        f'class {model.name}(Base):',
        f'    __tablename__ = "{model.table}"',
        '',
    ]

    for field in model.fields:
        col_type = type_to_python(field.type)
        constraints = []

        if 'pk' in field.constraints.lower():
            constraints.append('primary_key=True')
        if 'nullable' in field.constraints.lower():
            constraints.append('nullable=True')
        if 'default: now' in field.constraints.lower():
            constraints.append('server_default=func.now()')
        if 'fk:' in field.constraints.lower():
            # Extract foreign key reference
            import re
            fk_match = re.search(r'fk:\s*(\w+)', field.constraints.lower())
            if fk_match:
                constraints.insert(0, f'ForeignKey("{fk_match.group(1)}")')

        constraint_str = ', '.join(constraints) if constraints else ''
        if constraint_str:
            lines.append(f'    {field.name} = Column({col_type}, {constraint_str})')
        else:
            lines.append(f'    {field.name} = Column({col_type})')

    lines.extend([
        '',
        '    # TODO: Add relationships if needed',
        '',
    ])

    return '\n'.join(lines)


def generate_service_python(feature: str, services: List[FDLService], output_dir: str) -> str:
    """Generate Python service module."""
    lines = [
        '"""',
        f'{feature} Service',
        'Auto-generated from FDL. DO NOT modify function signatures.',
        '"""',
        'from typing import Any, Dict, List, Optional',
        'from uuid import UUID',
        '',
        '',
    ]

    for svc in services:
        # Build function signature
        input_params = []
        if svc.input:
            for param, ptype in svc.input.items():
                py_type = 'Any'
                if isinstance(ptype, str):
                    if ptype.lower() in ('uuid', 'string', 'str'):
                        py_type = 'str'
                    elif ptype.lower() in ('int', 'integer'):
                        py_type = 'int'
                    elif ptype.lower() in ('bool', 'boolean'):
                        py_type = 'bool'
                    elif ptype.lower() in ('list', 'array'):
                        py_type = 'List[Any]'
                    elif ptype.lower() in ('dict', 'object'):
                        py_type = 'Dict[str, Any]'
                input_params.append(f'{param}: {py_type}')

        params_str = ', '.join(input_params) if input_params else ''
        return_type = svc.output if svc.output else 'Any'

        lines.append(f'async def {svc.name}({params_str}) -> {return_type}:')
        lines.append(f'    """')
        lines.append(f'    {svc.desc}')
        lines.append(f'')
        lines.append(f'    Steps (from FDL):')
        for step in svc.steps:
            lines.append(f'    - {step}')
        lines.append(f'    """')
        lines.append(f'    # TODO: Implement the steps above')
        lines.append(f'    raise NotImplementedError("{svc.name} not implemented")')
        lines.append(f'')
        lines.append(f'')

    return '\n'.join(lines)


def generate_api_python(feature: str, apis: List[FDLAPI], output_dir: str) -> str:
    """Generate Python FastAPI router."""
    lines = [
        '"""',
        f'{feature} API Router',
        'Auto-generated from FDL. DO NOT modify endpoint signatures.',
        '"""',
        'from fastapi import APIRouter, HTTPException, Depends',
        'from pydantic import BaseModel',
        'from typing import Any, Dict, List, Optional',
        f'from app.services import {feature.lower()}_service',
        '',
        '',
        f'router = APIRouter(prefix="/{feature.lower()}", tags=["{feature}"])',
        '',
        '',
    ]

    for api in apis:
        method = api.method.lower()
        path = api.path

        # Generate request model if needed
        if api.request:
            class_name = f'{api.method.title()}{path.replace("/", "_").replace("{", "").replace("}", "").title()}Request'
            lines.append(f'class {class_name}(BaseModel):')
            for field, ftype in api.request.items():
                py_type = 'Any'
                if isinstance(ftype, str):
                    py_type = ftype
                lines.append(f'    {field}: {py_type}')
            lines.append('')
            lines.append('')

        # Generate endpoint
        decorator = f'@router.{method}("{path}"'
        if api.response:
            decorator += ', response_model=Dict[str, Any]'
        decorator += ')'
        lines.append(decorator)

        func_name = f'{method}_{path.replace("/", "_").replace("{", "").replace("}", "").strip("_")}'
        lines.append(f'async def {func_name}():')
        lines.append(f'    """')
        if api.summary:
            lines.append(f'    {api.summary}')
        lines.append(f'    Uses: {api.use}')
        lines.append(f'    """')
        if api.use:
            svc_func = api.use.replace('service.', '')
            lines.append(f'    # TODO: Call {feature.lower()}_service.{svc_func}()')
        lines.append(f'    raise NotImplementedError("API handler not implemented")')
        lines.append('')
        lines.append('')

    return '\n'.join(lines)


def generate_ui_react(feature: str, uis: List[FDLUI], output_dir: str) -> Dict[str, str]:
    """Generate React TSX components."""
    files = {}

    for ui in uis:
        lines = [
            f'/**',
            f' * {ui.component}',
            f' * Type: {ui.type}',
            f' * Auto-generated from FDL.',
            f' */',
            f'import React, {{ useState, useEffect }} from "react";',
            f'',
            f'',
        ]

        # Generate props interface
        if ui.props:
            lines.append(f'interface {ui.component}Props {{')
            for prop, ptype in ui.props.items():
                ts_type = 'any'
                if isinstance(ptype, str):
                    if ptype.lower() in ('string', 'str'):
                        ts_type = 'string'
                    elif ptype.lower() in ('number', 'int', 'integer', 'float'):
                        ts_type = 'number'
                    elif ptype.lower() in ('bool', 'boolean'):
                        ts_type = 'boolean'
                    elif ptype.lower() in ('array', 'list'):
                        ts_type = 'any[]'
                lines.append(f'  {prop}: {ts_type};')
            lines.append('}')
            lines.append('')
            lines.append('')

        # Generate component
        props_type = f'{ui.component}Props' if ui.props else '{}'
        lines.append(f'export const {ui.component}: React.FC<{props_type}> = (props) => {{')

        # Generate state
        for state in ui.state:
            if ':' in state:
                state_name, state_type = state.split(':', 1)
                state_name = state_name.strip()
                state_type = state_type.strip()
                if state_type.lower() == 'array':
                    lines.append(f'  const [{state_name}, set{state_name.title()}] = useState<any[]>([]);')
                elif state_type.lower() in ('string', 'str'):
                    lines.append(f'  const [{state_name}, set{state_name.title()}] = useState("");')
                elif state_type.lower() in ('number', 'int'):
                    lines.append(f'  const [{state_name}, set{state_name.title()}] = useState(0);')
                elif state_type.lower() in ('bool', 'boolean'):
                    lines.append(f'  const [{state_name}, set{state_name.title()}] = useState(false);')
                else:
                    lines.append(f'  const [{state_name}, set{state_name.title()}] = useState<any>(null);')
            else:
                lines.append(f'  const [{state}, set{state.title()}] = useState<any>(null);')

        lines.append('')

        # Generate useEffect for init
        if ui.init:
            lines.append('  useEffect(() => {')
            for init in ui.init:
                lines.append(f'    // TODO: {init}')
            lines.append('  }, []);')
            lines.append('')

        lines.extend([
            '  return (',
            '    <div>',
            f'      {{/* TODO: Implement {ui.component} UI */}}',
            '    </div>',
            '  );',
            '};',
            '',
        ])

        files[ui.component] = '\n'.join(lines)

    return files


def calculate_checksum(content: str) -> str:
    """Calculate SHA256 checksum of content."""
    return hashlib.sha256(content.encode()).hexdigest()


def write_file(path: str, content: str, force: bool) -> bool:
    """Write content to file."""
    if os.path.exists(path) and not force:
        return False

    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content)
    return True


def generate_all(
    spec: FDLSpec,
    output_dir: str,
    backend: str,
    frontend: str,
    force: bool,
    dry_run: bool
) -> Dict[str, Any]:
    """Generate all skeleton files."""
    result = {
        "generated_files": [],
        "errors": []
    }

    feature_lower = spec.feature.lower().replace('-', '_')

    # Generate models
    for model in spec.models:
        if backend == 'python':
            content = generate_model_python(model, output_dir)
            path = os.path.join(output_dir, 'models', f'{model.name.lower()}.py')

            if dry_run:
                result["generated_files"].append({
                    "path": path,
                    "layer": "model",
                    "checksum": calculate_checksum(content)
                })
            else:
                if write_file(path, content, force):
                    result["generated_files"].append({
                        "path": path,
                        "layer": "model",
                        "checksum": calculate_checksum(content)
                    })
                else:
                    result["errors"].append(f"File exists: {path}")

    # Generate services
    if spec.service:
        if backend == 'python':
            content = generate_service_python(spec.feature, spec.service, output_dir)
            path = os.path.join(output_dir, 'services', f'{feature_lower}_service.py')

            if dry_run:
                result["generated_files"].append({
                    "path": path,
                    "layer": "service",
                    "checksum": calculate_checksum(content)
                })
            else:
                if write_file(path, content, force):
                    result["generated_files"].append({
                        "path": path,
                        "layer": "service",
                        "checksum": calculate_checksum(content)
                    })
                else:
                    result["errors"].append(f"File exists: {path}")

    # Generate APIs
    if spec.api:
        if backend == 'python':
            content = generate_api_python(spec.feature, spec.api, output_dir)
            path = os.path.join(output_dir, 'api', f'{feature_lower}_api.py')

            if dry_run:
                result["generated_files"].append({
                    "path": path,
                    "layer": "api",
                    "checksum": calculate_checksum(content)
                })
            else:
                if write_file(path, content, force):
                    result["generated_files"].append({
                        "path": path,
                        "layer": "api",
                        "checksum": calculate_checksum(content)
                    })
                else:
                    result["errors"].append(f"File exists: {path}")

    # Generate UIs
    if spec.ui and frontend != 'none':
        if frontend == 'react':
            ui_files = generate_ui_react(spec.feature, spec.ui, output_dir)
            for component_name, content in ui_files.items():
                path = os.path.join(output_dir, 'components', f'{component_name}.tsx')

                if dry_run:
                    result["generated_files"].append({
                        "path": path,
                        "layer": "ui",
                        "checksum": calculate_checksum(content)
                    })
                else:
                    if write_file(path, content, force):
                        result["generated_files"].append({
                            "path": path,
                            "layer": "ui",
                            "checksum": calculate_checksum(content)
                        })
                    else:
                        result["errors"].append(f"File exists: {path}")

    return result


def main():
    parser = argparse.ArgumentParser(description='FDL Skeleton Generator')
    parser.add_argument('--fdl', required=True, help='FDL file path')
    parser.add_argument('--output-dir', default='.', help='Output directory')
    parser.add_argument('--backend', default='python', choices=['python', 'go', 'node'])
    parser.add_argument('--frontend', default='none', choices=['react', 'vue', 'none'])
    parser.add_argument('--force', action='store_true', help='Overwrite existing files')
    parser.add_argument('--dry-run', action='store_true', help='Show files that would be generated')

    args = parser.parse_args()

    try:
        spec = parse_fdl(args.fdl)
        result = generate_all(
            spec,
            args.output_dir,
            args.backend,
            args.frontend,
            args.force,
            args.dry_run
        )
        print(json.dumps(result, indent=2))
    except Exception as e:
        print(json.dumps({
            "generated_files": [],
            "errors": [str(e)]
        }))
        sys.exit(1)


if __name__ == '__main__':
    main()
