<template>
    <div class="dashboard-create-page">
        <!--        song-lang-->
        <p-page-title
            child
            title="Create New Dashboard"
            @goBack="$router.go(-1)"
        />
        <section class="dashboard-create-form-container">
            <dashboard-scope-form :dashboard-scope.sync="dashboardScope"
                                  @set-project="setForm('dashboardProject', $event)"
            />
            <dashboard-template-form @set-template="setForm('dashboardTemplate', $event)" />
            <dashboard-viewer-form :dashboard-viewer-type.sync="dashboardViewerType" />
        </section>
        <div class="dashboard-create-buttons">
            <p-button style-type="tertiary"
                      size="lg"
                      @click="$router.go(-1)"
            >
                <!--                song-lang-->
                Cancel
            </p-button>
            <p-button style-type="primary"
                      size="lg"
                      :disabled="!isAllValid"
                      @click="handleClickCreate"
            >
                <!--                song-lang-->
                Create
            </p-button>
        </div>
    </div>
</template>

<script lang="ts">
import { reactive, toRefs } from 'vue';

import { PPageTitle, PButton } from '@spaceone/design-system';

import { useFormValidator } from '@/common/composables/form-validator';

import {
    DASHBOARD_SCOPE,
    DASHBOARD_VIEWER,
} from '@/services/dashboards/config';
import DashboardScopeForm from '@/services/dashboards/dashboard-create/modules/DashboardScopeForm.vue';
import DashboardTemplateForm from '@/services/dashboards/dashboard-create/modules/DashboardTemplateForm.vue';
import DashboardViewerForm from '@/services/dashboards/dashboard-create/modules/DashboardViewerForm.vue';
import type { DashboardScope, DashboardViewer } from '@/services/dashboards/type';
import type { ProjectItemResp } from '@/services/project/type';

export default {
    name: 'CreateDashboardPage',
    components: {
        DashboardViewerForm,
        DashboardTemplateForm,
        DashboardScopeForm,
        PPageTitle,
        PButton,
    },
    setup() {
        const {
            forms: {
                dashboardTemplate,
                dashboardProject,
            },
            setForm,
            isAllValid,
        } = useFormValidator({
            dashboardTemplate: '',
            dashboardProject: undefined as ProjectItemResp|undefined,
        }, {
            // song-lang
            dashboardTemplate(value: boolean) { return !value ? 'Please Select Template' : ''; },
            dashboardProject(value: ProjectItemResp|undefined) {
                return !value && state.dashboardScope === DASHBOARD_SCOPE.PROJECT
                    // song-lang
                    ? 'Please Select Project' : '';
            },
        });

        const state = reactive({
            dashboardScope: DASHBOARD_SCOPE.DOMAIN as DashboardScope,
            dashboardViewerType: DASHBOARD_VIEWER.PUBLIC as DashboardViewer,
        });

        const handleClickCreate = () => {
            const dashboardCreateParams = {
                dashboardScope: state.dashboardScope,
                dashboardProject: state.dashboardScope === DASHBOARD_SCOPE.DOMAIN ? '' : dashboardProject.value,
                dashboardTemplate: dashboardTemplate.value,
                dashboardViewerType: state.dashboardViewerType,
            };
            console.log(dashboardCreateParams);
        };

        return {
            ...toRefs(state),
            dashboardTemplate,
            dashboardProject,
            setForm,
            isAllValid,
            handleClickCreate,
        };
    },
};
</script>

<style lang="postcss" scoped>
.dashboard-create-form-container {
    display: grid;
    grid-gap: 1rem;
}
.dashboard-create-buttons {
    display: inline-flex;
    float: right;
    margin-top: 1rem;
    & .p-button {
        margin-left: 1rem;
    }
}
</style>
