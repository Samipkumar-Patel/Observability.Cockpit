import React from 'react';
import { InlineField, Input } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from '../types';

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
    <InlineField label="Telemetry DB Path" labelWidth={20} tooltip="Path to the SQLite telemetry.db file that stores AI agent metrics.">
      <Input
        width={50}
        value={jsonData.dbPath || ''}
        onChange={handlePathChange}
        placeholder="telemetry.db"
      />
    </InlineField>
  );
}
