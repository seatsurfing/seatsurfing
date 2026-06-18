export default class EVENT {
  public static APPROVAL_COUNT_CHANGED = "approvalCountChanged";

  public static ApprovalCountChanged = () => {
    return new CustomEvent(this.APPROVAL_COUNT_CHANGED);
  };
}
