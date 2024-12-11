import { isAdministrator, useAuth } from '@/useAuth';
// import MilestonesBarChart from '@/Components/MilestonesBarChart';
// import ActivityChart from '@/Components/MonthActivityChart';
import StatsCard from '@/Components/StatsCard';
// import TopProgPieChart from '@/Components/TopProgActivityPieChart';
import { AdminLayer2Join, Facility, ServerResponse } from '@/common';
import useSWR from 'swr';
// import convertSeconds from '@/Components/ConvertSeconds';
// import { useContext } from 'react';
// import { ThemeContext } from '@/Context/ThemeContext';
import { AxiosError } from 'axios';
import UnauthorizedNotFound from './Unauthorized';
import { useEffect, useState } from 'react';
import API from '@/api/api';

export default function AdminLayer2() {
    const { user } = useAuth();
    const [facilities, setFacilities] = useState<Facility[]>();
    const [facility, setFacility] = useState('all');
    const [resetCache, setResetCache] = useState(false);

    const { data, error, isLoading } = useSWR<
        ServerResponse<AdminLayer2Join>,
        AxiosError
    >(`/api/users/${user?.id}/admin-layer2`);
    // const { theme } = useContext(ThemeContext);

    if (error || isLoading || !user) return <div></div>;
    if (!isAdministrator(user)) {
        return <UnauthorizedNotFound which="unauthorized" />;
    }

    // useEffect(() => {
    //     void mutate();
    // }, [facility, resetCache]);

    useEffect(() => {
        const fetchFacilities = async () => {
            const response = await API.get<Facility>('facilities');
            setFacilities(response.data as Facility[]);
        };
        void fetchFacilities();
    }, []);
    const metrics = data?.data;

    return (
        <div className="p-8">
            {error && <div>Error loading data</div>}
            {!data || (isLoading && <div>Loading...</div>)}
            {data && metrics && (
                <>
                    <div className="p-4">
                        <button
                            className="button"
                            onClick={() => setResetCache(!resetCache)}
                        >
                            Refresh Data
                        </button>
                        <div className="flex flex-row gap-4">
                            <div>
                                <label htmlFor="facility" className="label">
                                    <span className="label-text">Facility</span>
                                </label>
                                <select
                                    id="facility"
                                    className="select select-bordered w-full max-w-xs"
                                    value={facility}
                                    onChange={(e) =>
                                        setFacility(e.target.value)
                                    }
                                >
                                    <option key={'all'} value={'all'}>
                                        All
                                    </option>
                                    {facilities?.map((facility) => (
                                        <option
                                            key={facility.id}
                                            value={facility.id}
                                        >
                                            {facility.name}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </div>
                    <div className="grid grid-cols-3 gap-4 mb-6">
                        <StatsCard
                            title="Total Courses Offered"
                            number={'5'}
                            label="total_courses"
                        />
                        {/* <StatsCard
                            title="Active Users"
                            number={metrics.active_users.toString()}
                            label={`${(
                                (metrics.active_users / metrics.total_users) *
                                100
                            ).toFixed(2)}% of total`}
                        />
                        <StatsCard
                            title="Inactive Users"
                            number={(
                                metrics.total_users - metrics.active_users
                            ).toString()}
                            label="users"
                        /> */}
                    </div>
                    <div className="card card-row-padding mb-30">
                        {/* <div className="flex flex-row gap-6">
                            <div className="flex-1">
                                <ResponsiveContainer width="100%" height={300}>
                                    <PeakLoginTimesChart
                                        peak_login_times={
                                            metrics?.peak_login_times
                                        }
                                    />
                                </ResponsiveContainer>
                            </div>
                            <div className="flex-1">
                                <ResponsiveContainer width="100%" height={350}>
                                    <EngagementRateGraph
                                        active={metrics?.active_users}
                                        inactive={
                                            metrics.total_users -
                                            metrics.active_users
                                        }
                                    />
                                </ResponsiveContainer>
                            </div>
                        </div> */}
                    </div>
                </>
            )}
        </div>
    );
}
