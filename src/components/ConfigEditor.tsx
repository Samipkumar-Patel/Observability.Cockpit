import React from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from '../types';

const { FormField } = LegacyForms;

export interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { options, onOptionsChange } = props;
  const jsonData = options.jsonData || {};

  const handlePathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        dbPath: event.target.value,
      },
    });
  };

  return (
    <div className="gf-form-group">
      <div className="gf-form">
        <FormField
          label="Telemetry DB Path"
          labelWidth={10}
          inputWidth={25}
          onChange={handlePathChange}
          value={jsonData.dbPath || 'telemetry.db'}
          placeholder="Path to SQLite telemetry.db"
          tooltip="Path where AI agent token counts and metrics are stored."
        />
      </div>
    </div>
  );
}
