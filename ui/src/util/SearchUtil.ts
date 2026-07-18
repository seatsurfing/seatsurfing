import DateUtil from "@/util/DateUtil";

export default class SearchUtil {
  /**
   * Calculates the new enter/leave dates and the resulting
   * "autoUpdateEnterTimeToPrefWorkdayStart" flag for the search page's date
   * pickers, in response to a user changing the enter and/or leave date.
   *
   * @param currentEnter currently selected enter date
   * @param currentLeave currently selected leave date
   * @param selectionMultiDay whether a multi-day selection is currently allowed
   * @param dailyBasisBooking whether the org books on a daily basis
   * @param autoUpdateEnterTimeToPrefWorkdayStart current state of the "auto update enter time" flag
   * @param prefWorkdayStart user's preferred workday start time, format "HH:MM"
   * @param enter new enter date, or null if the enter date should remain unchanged
   * @param leave new leave date, or null if the leave date should remain unchanged
   * @returns the new enter/leave dates and the resulting auto-update flag, or
   *          null if nothing actually changed (caller should skip the update)
   */
  static calculateNewEnterAndLeave(
    currentEnter: Date,
    currentLeave: Date,
    selectionMultiDay: boolean,
    dailyBasisBooking: boolean,
    autoUpdateEnterTimeToPrefWorkdayStart: boolean,
    prefWorkdayStart: string,
    enter: Date | null,
    leave: Date | null,
  ): {
    newEnter?: Date;
    newLeave?: Date;
    autoUpdateEnterTimeToPrefWorkdayStart: boolean;
  } | null {
    if (enter === null && leave === null) return null;

    let newEnter: Date | undefined, newLeave: Date | undefined;

    // enter and leave change
    if (enter !== null && leave !== null) {
      if (
        DateUtil.equal(enter, currentEnter) &&
        DateUtil.equal(leave, currentLeave)
      )
        return null;
      newEnter = enter;
      newLeave = leave;

      // only enter change
    } else if (enter !== null) {
      if (DateUtil.equal(enter, currentEnter)) return null;
      newEnter = enter;
      const diff = currentLeave.getTime() - currentEnter.getTime();
      newLeave = new Date();
      newLeave.setTime(enter.getTime() + diff);
      if (!selectionMultiDay && !DateUtil.isSameDay(newLeave, enter)) {
        newLeave = DateUtil.setHoursToMax(new Date(enter));
      }

      // only leave change
    } else if (leave !== null) {
      if (DateUtil.equal(leave, currentLeave)) return null;
      newLeave = leave;
    }

    if (dailyBasisBooking) {
      if (newEnter) newEnter = DateUtil.setHoursToMin(newEnter);
      if (newLeave) newLeave = DateUtil.setHoursToMax(newLeave);
    } else if (autoUpdateEnterTimeToPrefWorkdayStart) {
      if (
        newEnter &&
        DateUtil.isSameTime(newEnter, currentEnter) &&
        !DateUtil.isSameDay(newEnter, currentEnter)
      ) {
        // enter date changed and enter time remains unchanged but -> set enter time to preferred time or next possible time
        if (DateUtil.isAfterToday(newEnter)) {
          newEnter = DateUtil.setTimeFromTimeString(newEnter, prefWorkdayStart);
        } else {
          newEnter = DateUtil.setTimeFromMinutes(
            newEnter,
            Math.max(
              DateUtil.timeStringToMinutes(prefWorkdayStart),
              (new Date().getHours() + 1) * 60,
            ),
          );
        }
      } else if (
        (newEnter && !DateUtil.isSameTime(newEnter, currentEnter)) ||
        (newLeave && !DateUtil.isSameTime(newLeave, currentLeave))
      ) {
        // user changed time -> no longer auto update time to preferred time
        autoUpdateEnterTimeToPrefWorkdayStart = false;
      }
    }

    return { newEnter, newLeave, autoUpdateEnterTimeToPrefWorkdayStart };
  }
}
