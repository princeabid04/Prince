<template>
    <section>
        <p-pane-layout>
            <!--            song-lang-->
            <p-panel-top title="Dashboard Scope" />
            <div class="dashboard-scope-wrapper">
                <p-radio-group direction="vertical">
                    <p-radio :selected="isDomainScope"
                             @change="handleSelectScope(DASHBOARD_SCOPE.DOMAIN)"
                    >
                        <!--                    song-lang-->
                        Entire Workspaces
                    </p-radio>
                    <p-radio :selected="!isDomainScope"
                             @change="handleSelectScope(DASHBOARD_SCOPE.PROJECT)"
                    >
                        <!--                    song-lang-->
                        Single Project
                    </p-radio>
                </p-radio-group>
                <project-select-dropdown v-show="!isDomainScope"
                                         project-selectable
                                         @select="handleSelectProjects"
                />
            </div>
        </p-pane-layout>
    </section>
</template>

<script lang="ts">
import type { SetupContext } from 'vue';
import { defineComponent, reactive, toRefs } from 'vue';

import {
    PPaneLayout, PPanelTop, PRadio, PRadioGroup,
} from '@spaceone/design-system';

import { store } from '@/store';

import ProjectSelectDropdown from '@/common/modules/project/ProjectSelectDropdown.vue';

import { DASHBOARD_SCOPE } from '@/services/dashboards/config';
import type { DashboardScope } from '@/services/dashboards/type';
import type { ProjectItemResp } from '@/services/project/type';

export default defineComponent({
    name: 'DashboardScopeForm',
    components: {
        ProjectSelectDropdown,
        PRadioGroup,
        PRadio,
        PPanelTop,
        PPaneLayout,
    },
    setup(props, { emit }: SetupContext) {
        const state = reactive({
            isDomainScope: true,
        });

        const handleSelectScope = (scopeType: DashboardScope) => {
            state.isDomainScope = scopeType === DASHBOARD_SCOPE.DOMAIN;
            emit('update:dashboardScope', scopeType);
        };

        const handleSelectProjects = (projects: Array<ProjectItemResp>) => {
            // Emit projects as project.
            emit('set-project', projects[0]);
        };

        // LOAD REFERENCE STORE
        (async () => {
            await store.dispatch('reference/project/load');
        })();

        return {
            ...toRefs(state),
            handleSelectScope,
            handleSelectProjects,
            DASHBOARD_SCOPE,
        };
    },
});
</script>

<style lang="postcss" scoped>
.dashboard-scope-wrapper {
    .p-radio-group {
        display: grid;
        grid-gap: 0.875rem;
    }
    .project-select-dropdown {
        margin: 0.375rem 0 0 1.125rem;
    }
    margin: 0.5rem 1rem 2.25rem;
}
</style>
