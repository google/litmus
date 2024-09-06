// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// icons
import {
  ExperimentOutlined,
  AppstoreAddOutlined,
  PlayCircleOutlined,
  LineChartOutlined,
  DeploymentUnitOutlined,
  QuestionOutlined,
  DatabaseOutlined
} from '@ant-design/icons-vue';

export interface menu {
  header?: string;
  title?: string;
  icon?: object;
  to?: string;
  divider?: boolean;
  chip?: string;
  chipColor?: string;
  chipVariant?: string;
  chipIcon?: string;
  children?: menu[];
  disabled?: boolean;
  type?: string;
  subCaption?: string;
}

const sidebarItem: menu[] = [
  { header: 'Testing' },
  {
    title: 'Tests',
    icon: ExperimentOutlined,
    to: '/'
  },
  {
    title: 'Compare Tests',
    icon: LineChartOutlined,
    to: '/compare-list'
  },
  {
    title: 'Start New Run',
    icon: PlayCircleOutlined,
    to: '/start'
  },
  { header: 'Configuration' },
  {
    title: 'Templates',
    icon: AppstoreAddOutlined,
    to: '/templates'
  },
  {
    title: 'Proxies',
    icon: DeploymentUnitOutlined,
    to: '/proxies'
  },
  {
    title: 'Help & FAQ',
    icon: QuestionOutlined,
    type: 'external',
    to: 'https://google.github.io/litmus/'
  }
];

export default sidebarItem;
