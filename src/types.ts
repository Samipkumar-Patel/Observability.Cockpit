import { DataSourceJsonData, DataQuery } from '@grafana/data';

export interface MyQuery extends DataQuery {
  queryText?: string;
  metric?: string;
}

export interface MyDataSourceOptions extends DataSourceJsonData {
  dbPath?: string;
}
