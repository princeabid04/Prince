<template>
    <div class="dashboard-board-list">
        <p-field-title class="p-field-title">
            <template>
                <span>{{ fieldTitle }}</span>
                <span class="board-count">({{ dashboardList.length }})</span>
            </template>
        </p-field-title>
        <p-board :board-sets="dashboardListByBoardSets"
                 class="board"
        >
            <template #item-content="{board}">
                <div class="board-item-title-wrapper">
                    <div class="favorite-button-wrapper">
                        <!--TODO: implementation about id-->
                        <favorite-button :item-id="board[`${dashboardScopeType}_dashboard_id`]"
                                         :favorite-type="FAVORITE_TYPE.DASHBOARD"
                                         scale="0.666"
                        />
                    </div>
                    <span class="board-item-title">{{ board.name }}</span>
                </div>
                <div class="board-item-description">
                    <span>{{ board.user_id }}</span>
                    <p-i name="ic_divider-dot"
                         width="0.125rem"
                         height="0.125rem"
                    />
                    <span>{{ dashboardScopeTypeForView }}</span>
                </div>
                <div class="label-wrapper">
                    <p-label :class="{'viewers-label': true, 'private-label': board.viewers === DASHBOARD_VIEWER.PRIVATE}"
                             :text="board.viewers === DASHBOARD_VIEWER.PUBLIC ? 'Public' : 'Private'"
                             :left-icon="board.viewers === DASHBOARD_VIEWER.PUBLIC ? 'ic_public' : 'ic_private'"
                    />
                    <p-label v-for="(label, idx) in board.labels"
                             :key="`${board.name}-label-${idx}`"
                             :text="label"
                             clickable
                             @item-click="handleSetQuery(label)"
                    />
                </div>
            </template>
        </p-board>
        <div v-if="dashboardList.length >= 10"
             class="dashboard-list-pagination"
        >
            <p-pagination :total-count="dashboardList.length"
                          :page-size="PAGE_SIZE"
                          :current-page="thisPage"
                          @change="handlePage"
            />
        </div>
        <delete-modal :header-title="deleteModalState.headerTitle"
                      :visible.sync="deleteModalState.visible"
                      :loading="deleteModalState.loading"
                      @confirm="handleDeleteDashboardConfirm"
        />
    </div>
</template>

<script lang="ts">
import type { PropType } from 'vue';
import {
    computed, defineComponent, reactive, toRefs, watch,
} from 'vue';

import {
    PBoard, PFieldTitle, PI, PLabel, PPagination,
} from '@spaceone/design-system';

import { QueryHelper } from '@cloudforet/core-lib/query';
import { SpaceConnector } from '@cloudforet/core-lib/space-connector';

import type { ConsoleFilter } from '@/query/type';
import { SpaceRouter } from '@/router';
import { store } from '@/store';
import { i18n } from '@/translations';

import type { DashboardModel } from '@/store/modules/dashboard/type';
import { FAVORITE_TYPE } from '@/store/modules/favorite/type';

import { showSuccessMessage } from '@/lib/helper/notice-alert-helper';

import DeleteModal from '@/common/components/modals/DeleteModal.vue';
import ErrorHandler from '@/common/composables/error/errorHandler';
import FavoriteButton from '@/common/modules/favorites/favorite-button/FavoriteButton.vue';

import { DASHBOARD_SCOPE, DASHBOARD_VIEWER } from '@/services/dashboards/config';
import { DASHBOARDS_ROUTE } from '@/services/dashboards/route-config';
import type { DashboardScope } from '@/services/dashboards/type';

const PAGE_SIZE = 10;

interface DashboardBoardListProps {
    scopeType: DashboardScope;
    fieldTitle: string;
    // TODO: implementation
    dashboardList: DashboardModel[];
}

export default defineComponent<DashboardBoardListProps>({
    name: 'DashboardBoardList',
    components: {
        PI,
        DeleteModal,
        PPagination,
        PLabel,
        FavoriteButton,
        PBoard,
        PFieldTitle,
    },
    props: {
        scopeType: {
            type: String as PropType<DashboardScope>,
            default: undefined,
        },
        fieldTitle: {
            type: String,
            default: undefined,
        },
        dashboardList: {
            type: Array as PropType<DashboardModel[]>,
            default: () => [],
        },
    },
    setup(props) {
        /* song-lang */
        const state = reactive({
            thisPage: 1,
            dashboardScopeType: computed(() => (props.scopeType === DASHBOARD_SCOPE.DOMAIN ? 'domain' : 'project')),
            dashboardListByBoardSets: computed(() => props.dashboardList
                .slice((state.thisPage - 1) * PAGE_SIZE, state.thisPage * PAGE_SIZE)
                .map((d) => (
                    {
                        ...d,
                        iconButtonSets: convertBoardItemButtonSet(d[`${state.dashboardScopeType}_dashboard_id`]),
                    }
                ))),
        });

        /* song-lang */
        const deleteModalState = reactive({
            headerTitle: i18n.t('Are you sure you want to delete dashboard?'),
            visible: false,
            loading: false,
            selectedId: undefined as string|undefined,
        });

        const handleClickDeleteDashboard = (dashboardId) => {
            deleteModalState.selectedId = dashboardId;
            deleteModalState.visible = true;
        };

        const handleDeleteDashboardConfirm = async () => {
            try {
                deleteModalState.loading = true;
                if (state.dashboardScopeType === 'domain') {
                    await SpaceConnector.clientV2.dashboard.domainDashboard.delete({
                        domain_dashboard_id: deleteModalState.selectedId,
                    });
                    await store.dispatch('dashboard/loadDomainDashboard');
                } else {
                    await SpaceConnector.clientV2.dashboard.projectDashboard.delete({
                        project_dashboard_id: deleteModalState.selectedId,
                    });
                    await store.dispatch('dashboard/loadProjectDashboard');
                }
                /* song-lang */
                showSuccessMessage(i18n.t('Successed to delete dashboard'), '');
            } catch (e) {
                /* song-lang */
                ErrorHandler.handleRequestError(e, i18n.t('Failed to delete dashboard'));
            } finally {
                deleteModalState.loading = false;
                deleteModalState.visible = false;
                deleteModalState.selectedId = undefined;
            }
        };

        /* song-lang */
        const convertBoardItemButtonSet = (dashboardId) => [
            {
                iconName: 'ic_edit',
                tooltipText: i18n.t('Edit'),
                eventAction: () => {
                    SpaceRouter.router.push({
                        name: DASHBOARDS_ROUTE.CUSTOMIZE._NAME,
                        params: { id: dashboardId },
                    });
                },
            },
            {
                iconName: 'ic_duplicate',
                tooltipText: i18n.t('Clone'),
                eventAction: () => console.log('dup!'),
            },
            {
                iconName: 'ic_trashcan',
                tooltipText: i18n.t('Delete'),
                /* TODO: Implementation */
                eventAction: () => handleClickDeleteDashboard(dashboardId),
            },
        ];

        const labelQueryHelper = new QueryHelper();
        const handleSetQuery = (selectedLabel: ConsoleFilter | ConsoleFilter[]) => {
            labelQueryHelper.setFilters(store.state.dashboard.searchFilters).addFilter({ k: 'label', o: '=', v: selectedLabel });
            store.dispatch('dashboard/setSearchFilters', labelQueryHelper.filters);
        };
        const handlePage = (page: number) => {
            state.thisPage = page;
        };

        watch(() => props.dashboardList, () => {
            state.thisPage = 1;
        });

        return {
            ...toRefs(state),
            deleteModalState,
            handleSetQuery,
            handlePage,
            handleDeleteDashboardConfirm,
            FAVORITE_TYPE,
            DASHBOARD_SCOPE,
            DASHBOARD_VIEWER,
            PAGE_SIZE,
        };
    },
});
</script>

<style lang="postcss" scoped>
.dashboard-board-list {
    flex-grow: 1;
    margin-top: 0.5rem;

    /* custom design-system component - p-field-title */
    :deep(.p-field-title) {
        margin-bottom: 0.5rem;
    }
    .board-count {
        font-weight: normal;
        margin-left: 0.5rem;
    }
    .board {
        .board-item-title-wrapper {
            @apply flex items-center;
            height: 1.25rem;
            .favorite-button-wrapper {
                @apply flex items-center justify-center;
                width: 1.25rem;
                height: 1.25rem;
            }
            .board-item-title {
                margin-left: 0.125rem;
                font-size: 1rem;
                font-weight: bold;
                line-height: 1.25;
            }
        }
        .board-item-description {
            @apply flex items-center;
            gap: 0.5rem;
            font-size: 0.75rem;
            line-height: 1.25;
            color: gray;
            margin: 0.25rem 0 0.75rem;
        }

        .label-wrapper {
            @apply flex items-center;
        }

        /* custom design-system component - p-label */
        :deep(.viewers-label) {
            @apply border-0 bg-violet-200;
            &.private-label {
                @apply bg-gray-200;
            }
        }
    }
    .dashboard-list-pagination {
        @apply w-full flex justify-center;
        margin-top: 0.5rem;
    }
}
</style>
