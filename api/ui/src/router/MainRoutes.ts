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

const MainRoutes = {
  path: '/main',
  meta: {
    requiresAuth: true
  },
  redirect: '/main',
  component: () => import('@/layouts/common/Layout.vue'),
  children: [
    {
      path: '/',
      name: 'tests',
      component: () => import('@/views/TestListView.vue'),
      meta: { title: 'Tests' }
    },
    {
      path: '/start',
      name: 'submitrun',
      component: () => import('@/views/SubmitRunView.vue'),
      meta: { title: 'SubmitRun' }
    },
    {
      path: '/templates',
      name: 'templates',
      component: () => import('@/views/TemplateListView.vue'),
      meta: { title: 'Templates' }
    },
    {
      path: '/proxies',
      name: 'proxies',
      component: () => import('@/views/ProxiesListView.vue'),
      meta: { title: 'Proxies' }
    },
    {
      path: '/compare-list',
      name: 'compare-list',
      component: () => import('@/views/CompareListView.vue'),
      meta: { title: 'Compare List' }
    },
    {
      path: '/compare/:templateId',
      name: 'compare',
      component: () => import('../views/CompareRuns.vue'),
      props: true // Allow passing templateId as a prop
    },
    {
      path: '/edit-template/:templateId',
      name: 'editTemplate',
      component: () => import('@/views/EditTemplatePage.vue'),
      props: true // Allow passing templateId as a prop
    },
    {
      path: '/add-template',
      name: 'addTemplate',
      component: () => import('@/views/AddTemplatePage.vue')
    },
    {
      path: '/runs/:runId', // Dynamic route for run details
      name: 'runDetails',
      component: () => import('../views/RunDetailsView.vue') // Lazy load the component
    },
    {
      path: '/data-explorer', // Dynamic route for run details
      name: 'dataExplorer',
      component: () => import('../views/DataExplorerView.vue') // Lazy load the component
    },
    {
      path: '/files', // Dynamic route for run details
      name: 'fileManager',
      component: () => import('../views/FileManagerView.vue') // Lazy load the component
    }
  ]
};

export default MainRoutes;
