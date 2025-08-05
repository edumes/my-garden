'use client';

import { Activity, ArrowUpRight, Leaf, ShoppingCart, TreePine, User, User2 } from 'lucide-react';
import { motion, type Transition } from 'motion/react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { toast } from 'sonner';
import { apiService } from '../services/api';

interface AuditLog {
  id: string;
  user_id?: string;
  action: string;
  resource: string;
  resource_id?: string;
  details: any;
  ip_address: string;
  user_agent: string;
  status: string;
  created_at: string;
  user?: {
    username: string;
    email: string;
  };
}

const transition: Transition = {
  type: 'spring',
  stiffness: 300,
  damping: 26,
};

const getCardVariants = (i: number) => ({
  collapsed: {
    marginTop: i === 0 ? 0 : -44,
    scaleX: 1 - i * 0.05,
  },
  expanded: {
    marginTop: i === 0 ? 0 : 4,
    scaleX: 1,
  },
});

const textSwitchTransition: Transition = {
  duration: 0.22,
  ease: 'easeInOut',
};

const notificationTextVariants = {
  collapsed: { opacity: 1, y: 0, pointerEvents: 'auto' },
  expanded: { opacity: 0, y: -16, pointerEvents: 'none' },
};

const viewAllTextVariants = {
  collapsed: { opacity: 0, y: 16, pointerEvents: 'none' },
  expanded: { opacity: 1, y: 0, pointerEvents: 'auto' },
};

const getActionIcon = (action: string) => {
  switch (action) {
    case 'user_register':
    case 'user_login':
    case 'user_logout':
      return <User className="size-3" />;
    case 'garden_create':
    case 'garden_update':
    case 'garden_delete':
    case 'garden_view':
      return <TreePine className="size-3" />;
    case 'plant_seed':
    case 'plant_harvest':
    case 'plant_remove':
      return <Leaf className="size-3" />;
    case 'seed_purchase':
      return <ShoppingCart className="size-3" />;
    default:
      return <Activity className="size-3" />;
  }
};

const getActionTitle = (action: string) => {
  switch (action) {
    case 'user_register':
      return 'User Registered';
    case 'user_login':
      return 'User Login';
    case 'user_logout':
      return 'User Logout';
    case 'garden_create':
      return 'Garden Created';
    case 'garden_update':
      return 'Garden Updated';
    case 'garden_delete':
      return 'Garden Deleted';
    case 'plant_seed':
      return 'Plant Seeded';
    case 'plant_harvest':
      return 'Plant Harvested';
    case 'plant_remove':
      return 'Plant Removed';
    case 'seed_purchase':
      return 'Seeds Purchased';
    default:
      return action.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }
};

const getActionSubtitle = (action: string, details: any, status: string) => {
  switch (action) {
    case 'user_register':
    case 'user_login':
      return details?.username ? `by ${details.username}` : 'User action';
    case 'garden_create':
      return details?.garden_name ? `"${details.garden_name}"` : 'New garden';
    case 'plant_seed':
      return details?.plant_type_id ? `Plant type: ${details.plant_type_id}` : 'Plant seeded';
    case 'plant_harvest':
      return details?.coins_earned ? `+${details.coins_earned} coins` : 'Plant harvested';
    case 'seed_purchase':
      return details?.total_cost ? `${details.quantity} seeds for ${details.total_cost} coins` : 'Seeds purchased';
    default:
      return status === 'success' ? 'Action completed' : 'Action failed';
  }
};

const formatTimeAgo = (dateString: string) => {
  const now = new Date();
  const date = new Date(dateString);
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) {
    return 'just now';
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return `${minutes}m`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return `${hours}h`;
  } else {
    const days = Math.floor(diffInSeconds / 86400);
    return `${days}d`;
  }
};

function AuditNotifications() {
  const [notifications, setNotifications] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const loadRecentAuditLogs = async () => {
    setLoading(true);
    try {
      const response = await apiService.getAuditLogs({
        limit: 5,
        status: 'success'
      }) as any;

      const recentLogs = response.logs.map((log: AuditLog) => ({
        ...log,
        title: getActionTitle(log.action),
        subtitle: getActionSubtitle(log.action, log.details, log.status),
        time: formatTimeAgo(log.created_at),
        icon: getActionIcon(log.action),
      }));

      setNotifications(recentLogs);
    } catch (error) {
      console.error('Failed to load audit notifications:', error);
      toast.error('Failed to load recent activity');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRecentAuditLogs();

    const interval = setInterval(loadRecentAuditLogs, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="bg-neutral-200 dark:bg-neutral-900 p-3 rounded-3xl w-xs space-y-3 shadow-md">
        <div className="animate-pulse">
          <div className="bg-neutral-100 dark:bg-neutral-800 rounded-xl px-4 py-2 h-16"></div>
          <div className="bg-neutral-100 dark:bg-neutral-800 rounded-xl px-4 py-2 h-16 mt-2"></div>
        </div>
      </div>
    );
  }

  return (
    <motion.div
      className="bg-neutral-200 dark:bg-neutral-900 p-3 rounded-3xl w-xs space-y-3 shadow-md"
      initial="collapsed"
      whileHover="expanded"
    >
      <div>
        {notifications.map((notification, i) => (
          <motion.div
            key={notification.id}
            className="bg-neutral-100 dark:bg-neutral-800 rounded-xl px-4 py-2 shadow-sm hover:shadow-lg transition-shadow duration-200 relative"
            variants={getCardVariants(i)}
            transition={transition}
            style={{
              zIndex: notifications.length - i,
            }}
          >
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-2">
                {getActionIcon(notification.action)}
                <h1 className="text-sm font-medium">{getActionTitle(notification.action)}</h1>
              </div>
              <div className="flex items-center text-xs gap-0.5 font-medium text-neutral-500 dark:text-neutral-300">
                <User2 className="size-3" />
                <span>{notification.user?.username || 'Unknown'}</span>
              </div>
            </div>
            <div className="text-xs text-neutral-500 font-medium">
              <span>{formatTimeAgo(notification.created_at)}</span>
              &nbsp;•&nbsp;
              <span>{getActionSubtitle(notification.action, notification.details, notification.status)}</span>
            </div>
          </motion.div>
        ))}
      </div>

      <div className="flex items-center gap-2">
        <div className="size-5 rounded-full bg-neutral-400 text-white text-xs flex items-center justify-center font-medium">
          {notifications.length}
        </div>
        <span className="grid">
          <motion.span
            className="text-sm font-medium text-neutral-600 dark:text-neutral-300 row-start-1 col-start-1"
            variants={notificationTextVariants}
            transition={textSwitchTransition}
          >
            Recent Activity
          </motion.span>
          <motion.span
            className="text-sm font-medium text-neutral-600 dark:text-neutral-300 flex items-center gap-1 cursor-pointer select-none row-start-1 col-start-1"
            variants={viewAllTextVariants}
            transition={textSwitchTransition}
            onClick={() => navigate('/audit')}
          >
            View all <ArrowUpRight className="size-4" />
          </motion.span>
        </span>
      </div>
    </motion.div>
  );
}

export { AuditNotifications };
