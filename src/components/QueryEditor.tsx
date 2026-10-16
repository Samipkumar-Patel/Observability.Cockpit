import React from 'react';
import { InlineField, Input, Select } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyQuery, MyDataSourceOptions } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onMetricChange = (item: any) => {
    onChange({ ...query, metric: item.value });
    onRunQuery();
  };

  const onQueryTextChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, queryText: event.target.value });
  };

  const metricOptions = [
    { label: 'Token Consumption (tokens_consumed)', value: 'tokens_consumed' },
    { label: 'Prompt Latency (prompt_latency_ms)', value: 'prompt_latency_ms' },
  ];

  return (
    <div className="gf-form">
      <InlineField label="LLM Metric">
        <Select
          options={metricOptions}
          value={query.metric || 'tokens_consumed'}
          onChange={onMetricChange}
          width={30}
        />
      </InlineField>
      <InlineField label="Query Filter">
        <Input
          value={query.queryText || ''}
          onChange={onQueryTextChange}
          onBlur={onRunQuery}
          placeholder="e.g. agent_name = customer-support"
          width={40}
        />
      </InlineField>
    </div>
  );
}
