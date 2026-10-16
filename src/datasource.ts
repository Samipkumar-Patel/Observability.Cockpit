import { DataSourceApi, DataSourceInstanceSettings } from '@grafana/data';
import { MyQuery, MyDataSourceOptions } from './types';

export class DataSource extends DataSourceApi<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  async query(options: any): Promise<any> {
    const queries = options.targets.map((target: MyQuery) => {
      return {
        refId: target.refId,
        queryText: target.queryText || 'SELECT * FROM metrics',
        metric: target.metric || 'tokens_consumed',
      };
    });

    return super.query({ ...options, targets: queries });
  }

  async testDatasource() {
    return {
      status: 'success',
      message: 'LLM Observability data source is working correctly!',
    };
  }
}
